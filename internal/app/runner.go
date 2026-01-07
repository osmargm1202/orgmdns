package app

import (
	"fmt"
	"time"

	"github.com/osmargm1202/orgmdns/internal/cloudflare"
	"github.com/osmargm1202/orgmdns/internal/config"
	"github.com/osmargm1202/orgmdns/internal/ip"
	"github.com/osmargm1202/orgmdns/internal/logger"
	"github.com/osmargm1202/orgmdns/internal/notify"
)

type Runner struct {
	config          *config.Config
	logger          *logger.Logger
	cf              *cloudflare.Client
	notifier        *notify.EmailNotifier
	internetDown    bool
	disconnectedAt   *time.Time
	startupEmailSent bool
}

func NewRunner(cfg *config.Config, log *logger.Logger) *Runner {
	cfClient := cloudflare.NewClient(cfg.AccountID, cfg.APIKey, cfg.ZoneID, cfg.APIEmail)
	emailNotifier := notify.NewEmailNotifier(cfg.EmailFrom, cfg.EmailTo, cfg.EmailPassword, cfg.SMTPHost, cfg.SMTPPort)

	// Log del método de autenticación usado
	if cfg.APIEmail != "" {
		log.Info(fmt.Sprintf("Usando autenticación Cloudflare: API Key + Email (método legacy) - Email: %s", cfg.APIEmail))
	} else {
		log.Info(fmt.Sprintf("Usando autenticación Cloudflare: API Token (Bearer)"))
		log.Info(fmt.Sprintf("Si tienes problemas de autenticación, configura API_EMAIL con tu email de Cloudflare"))
	}

	return &Runner{
		config:   cfg,
		logger:   log,
		cf:       cfClient,
		notifier: emailNotifier,
	}
}

func (r *Runner) Run() error {
	r.logger.Info("Iniciando bucle principal de verificación de IP")

	for {
		r.logger.Debug("Iniciando ciclo de verificación")

		// Verificar conexión a internet (como en Python)
		if !ip.CheckInternetConnection() {
			if !r.internetDown {
				// Primera vez que se detecta sin conexión
				now := time.Now()
				r.disconnectedAt = &now
				r.internetDown = true
				r.logger.Error("No hay conexión a internet")
				// No enviamos correo aquí porque no hay conexión para enviarlo
			}
			r.sleep()
			continue
		}

		// Si llegamos aquí, hay conexión a internet
		if r.internetDown {
			// Se ha restaurado la conexión
			r.logger.Info("Se ha restaurado la conexión a internet")
			
			// Calcular tiempo sin conexión
			if r.disconnectedAt != nil {
				duration := time.Since(*r.disconnectedAt)
				r.logger.Info(fmt.Sprintf("Tiempo sin conexión: %v", duration))
				
				// Enviar correo de restauración
				if err := r.notifier.SendConnectionRestoredNotification(duration); err != nil {
					r.logger.Error(fmt.Sprintf("Error enviando correo de restauración: %v", err))
				} else {
					r.logger.Info("Correo de restauración de conexión enviado")
				}
			}
			
			r.internetDown = false
			r.disconnectedAt = nil
		}

		// Obtener IP pública actual
		currentIP, err := ip.GetPublicIP()
		if err != nil {
			r.logger.Error(fmt.Sprintf("Error obteniendo IP pública: %v", err))
			// Continuar al siguiente ciclo después del sleep
			r.sleep()
			continue
		}

		r.logger.Info(fmt.Sprintf("IP pública detectada: %s", currentIP))

		// Enviar correo de inicio solo la primera vez
		if !r.startupEmailSent {
			if err := r.notifier.SendStartupNotification(currentIP, r.config.RecordNames); err != nil {
				r.logger.Error(fmt.Sprintf("Error enviando correo de inicio: %v", err))
			} else {
				r.logger.Info("Correo de inicio enviado: Verificador DNS corriendo")
				r.startupEmailSent = true
			}
		}

		r.logger.Debug(fmt.Sprintf("Verificando %d registros DNS", len(r.config.RecordNames)))

		// Procesar cada registro
		for _, recordName := range r.config.RecordNames {
			if err := r.processRecord(recordName, currentIP); err != nil {
				r.logger.Error(fmt.Sprintf("Error procesando registro %s: %v", recordName, err))
				// Continuar con el siguiente registro
				continue
			}
		}

		r.logger.Debug(fmt.Sprintf("Ciclo completado, esperando %d minutos", r.config.SleepTime))
		r.sleep()
	}
}

func (r *Runner) processRecord(recordName, currentIP string) error {
	r.logger.Debug(fmt.Sprintf("Procesando registro: %s", recordName))

	// Obtener registro actual de Cloudflare (obtiene todos y filtra localmente como Python)
	record, err := r.cf.GetDNSRecordByName(recordName)
	if err != nil {
		return fmt.Errorf("error obteniendo registro DNS: %w", err)
	}

	r.logger.Debug(fmt.Sprintf("Registro DNS encontrado: %s -> %s (ID: %s)", record.Name, record.Content, record.ID))

	// Comparar IPs
	if record.Content == currentIP {
		r.logger.Debug(fmt.Sprintf("IP del registro %s coincide con IP actual (%s), no se requiere actualización", recordName, currentIP))
		return nil
	}

	oldIP := record.Content
	r.logger.Info(fmt.Sprintf("IP diferente detectada para %s: DNS=%s, Actual=%s. Actualizando...", recordName, oldIP, currentIP))

	// Actualizar registro en Cloudflare
	if err := r.cf.UpdateDNSRecordIP(record.ID, currentIP); err != nil {
		return fmt.Errorf("error actualizando registro DNS: %w", err)
	}

	r.logger.Info(fmt.Sprintf("Registro %s actualizado exitosamente: %s -> %s", recordName, oldIP, currentIP))

	// Enviar notificación por correo
	if err := r.notifier.SendDNSUpdateNotification(recordName, oldIP, currentIP); err != nil {
		r.logger.Error(fmt.Sprintf("Error enviando correo de notificación: %v", err))
		// No retornamos error aquí, el cambio de DNS ya se hizo
	} else {
		r.logger.Debug(fmt.Sprintf("Correo de notificación enviado para %s", recordName))
	}

	return nil
}

func (r *Runner) sleep() {
	duration := r.config.SleepDuration()
	r.logger.Debug(fmt.Sprintf("Durmiendo por %v", duration))
	time.Sleep(duration)
}
