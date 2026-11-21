package app

import (
	"fmt"
	"os"
	"time"

	"github.com/osmargm1202/orgmdns/internal/cloudflare"
	"github.com/osmargm1202/orgmdns/internal/config"
	"github.com/osmargm1202/orgmdns/internal/ip"
	"github.com/osmargm1202/orgmdns/internal/logger"
	"github.com/osmargm1202/orgmdns/internal/notify"
)

type dnsClient interface {
	GetDNSRecordByName(string) (*cloudflare.DNSRecord, error)
	UpdateDNSRecordIP(string, string) error
}
type summaryNotifier interface{ SendSummary([]notify.Event) error }

type Runner struct {
	config         *config.Config
	logger         *logger.Logger
	cf             dnsClient
	notifier       summaryNotifier
	internetDown   bool
	disconnectedAt *time.Time
	startupQueued  bool
	lastIP         string
	pending        []notify.Event
	checkInternet  func() bool
	publicIP       func() (string, error)
}

func NewRunner(cfg *config.Config, log *logger.Logger) *Runner {
	return &Runner{
		config: cfg, logger: log,
		cf:            cloudflare.NewClient(cfg.AccountID, cfg.APIKey, cfg.ZoneID, cfg.APIEmail),
		notifier:      notify.NewEmailNotifier(cfg.EmailFrom, cfg.EmailTo, cfg.EmailPassword, cfg.SMTPHost, cfg.SMTPPort),
		checkInternet: ip.CheckInternetConnection, publicIP: ip.GetPublicIP,
	}
}

func (r *Runner) Run() error {
	r.logger.Info("Iniciando bucle principal de verificación de IP")
	for {
		r.runCycle()
		time.Sleep(r.config.SleepDuration())
	}
}

func (r *Runner) addEvent(kind, record, oldIP, newIP, detail string) {
	// Coalesce identical pending failures while SMTP is unavailable.
	for _, e := range r.pending {
		if e.Kind == kind && e.Record == record && e.OldIP == oldIP && e.NewIP == newIP && e.Detail == detail {
			return
		}
	}
	event := notify.Event{Time: time.Now(), Kind: kind, Record: record, OldIP: oldIP, NewIP: newIP, Detail: detail}
	if r.config.NotificationDir != "" {
		if err := notify.Enqueue(r.config.NotificationDir, event); err == nil {
			return
		} else {
			r.logger.Error(fmt.Sprintf("No se pudo guardar evento; se conserva en memoria: %v", err))
		}
	}
	r.pending = append(r.pending, event)
}

func (r *Runner) runCycle() {
	if !r.checkInternet() {
		if !r.internetDown {
			now := time.Now()
			r.disconnectedAt = &now
			r.internetDown = true
			r.logger.Error("No hay conexión a internet")
		}
		return
	}
	if r.internetDown {
		detail := "Conexión restaurada"
		if r.disconnectedAt != nil {
			detail += fmt.Sprintf("; tiempo sin conexión: %s", time.Since(*r.disconnectedAt).Round(time.Second))
		}
		r.addEvent("Reconexión", "", "", "", detail)
		r.internetDown = false
		r.disconnectedAt = nil
	}
	currentIP, err := r.publicIP()
	if err != nil {
		r.logger.Error(fmt.Sprintf("Error obteniendo IP pública: %v", err))
		r.addEvent("Error", "", "", "", "No se pudo obtener la IP pública; no se modificaron registros DNS")
		r.flush()
		return
	}
	if !r.startupQueued {
		r.addEvent("Inicio", "", "", currentIP, "Verificador iniciado")
		r.startupQueued = true
	}
	if r.lastIP != "" && r.lastIP != currentIP {
		r.addEvent("Cambio de IP pública", "", r.lastIP, currentIP, "IP detectada; consultar el resultado DNS en las filas siguientes")
	}
	r.lastIP = currentIP
	for _, name := range r.config.RecordNames {
		if err := r.processRecord(name, currentIP); err != nil {
			r.logger.Error(fmt.Sprintf("Error procesando %s: %v", name, err))
		}
	}
	r.flush()
}

func (r *Runner) processRecord(name, currentIP string) error {
	record, err := r.cf.GetDNSRecordByName(name)
	if err != nil {
		r.addEvent("Error DNS", name, "", currentIP, "No se pudo consultar el registro; no se confirmó ninguna actualización")
		return err
	}
	if record.Content == currentIP {
		return nil
	}
	if err := r.cf.UpdateDNSRecordIP(record.ID, currentIP); err != nil {
		r.addEvent("Error DNS", name, record.Content, currentIP, "Cloudflare no confirmó la actualización; se reintentará en el próximo ciclo")
		return err
	}
	r.addEvent("DNS actualizado", name, record.Content, currentIP, "Actualización confirmada por Cloudflare")
	r.logger.Info(fmt.Sprintf("Registro %s actualizado: %s -> %s", name, record.Content, currentIP))
	return nil
}

func (r *Runner) flush() {
	events := append([]notify.Event(nil), r.pending...)
	var queued []notify.QueuedEvent
	if r.config.NotificationDir != "" {
		var err error
		queued, err = notify.ReadQueue(r.config.NotificationDir)
		if err != nil {
			r.logger.Error(fmt.Sprintf("No se pudo leer la cola: %v", err))
			return
		}
		for _, entry := range queued {
			events = append(events, entry.Event)
		}
	}
	if len(events) == 0 {
		return
	}
	// Wait for a quiet interval so the independent server monitor can contribute
	// reconnection/IP events before the DNS process sends the incident summary.
	for _, event := range events {
		if time.Since(event.Time) < r.config.NotificationSettle {
			return
		}
	}
	if err := r.notifier.SendSummary(events); err != nil {
		r.logger.Error(fmt.Sprintf("No se pudo enviar el resumen; se conserva para reintentar: %v", err))
		return
	}
	for _, entry := range queued {
		if err := os.Remove(entry.Path); err != nil {
			r.logger.Error(fmt.Sprintf("Resumen enviado pero no se pudo confirmar evento en cola: %v", err))
		}
	}
	r.logger.Info(fmt.Sprintf("Resumen enviado con %d eventos", len(events)))
	r.pending = nil
}
