package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/smtp"
	"net/textproto"
	"time"
)

type Event struct {
	Time   time.Time
	Kind   string
	Record string
	OldIP  string
	NewIP  string
	Detail string
}

type EmailNotifier struct {
	from, to, password, smtpHost, smtpPort string
}

func NewEmailNotifier(from, to, password, smtpHost, smtpPort string) *EmailNotifier {
	return &EmailNotifier{from, to, password, smtpHost, smtpPort}
}

var summaryTemplate = template.Must(template.New("summary").Parse(`<!doctype html>
<html lang="es"><body style="margin:0;background:#f1f5f9;font-family:Arial,sans-serif;color:#1e293b">
<table role="presentation" width="100%"><tr><td style="padding:24px">
<table role="presentation" width="100%" style="max-width:900px;margin:auto;background:white;border-radius:12px"><tr><td style="padding:28px;background:#0f172a;color:white">
<p style="margin:0;color:#93c5fd">ORGM · INFRAESTRUCTURA</p><h1 style="font-size:24px">Resumen de servidor y DNS</h1>
<p>{{.Count}} eventos · {{.Date}}</p></td></tr><tr><td style="padding:24px">
<p>Resultado de la verificación automática. Las actualizaciones confirmadas y los errores aparecen por separado en cada fila.</p>
<table width="100%" cellpadding="10" cellspacing="0" style="border-collapse:collapse;font-size:13px;text-align:left">
<tr style="background:#e2e8f0"><th>Fecha / evento</th><th>Registro</th><th>IP anterior</th><th>IP nueva</th><th>Resultado</th></tr>
{{range .Events}}<tr><td style="border-bottom:1px solid #e2e8f0">{{.Time.Format "2006-01-02 15:04:05 MST"}}<br><strong>{{.Kind}}</strong></td><td>{{.Record}}</td><td>{{.OldIP}}</td><td>{{.NewIP}}</td><td style="border-bottom:1px solid #e2e8f0;white-space:pre-line">{{.Detail}}</td></tr>{{end}}
</table><p style="color:#64748b;font-size:12px">Aviso automático de orgmdns. Un resumen por ciclo con novedades; los ciclos sin novedades no generan correo.</p>
</td></tr></table></td></tr></table></body></html>`))

// BuildMessage creates a UTF-8 multipart email; html/template escapes all event data.
func BuildMessage(from, to string, events []Event) ([]byte, error) {
	var html bytes.Buffer
	data := struct {
		Count  int
		Date   string
		Events []Event
	}{len(events), time.Now().Format(time.RFC3339), events}
	if err := summaryTemplate.Execute(&html, data); err != nil {
		return nil, err
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	var plain bytes.Buffer
	fmt.Fprintln(&plain, "ORGM — Resumen de servidor y DNS")
	for _, event := range events {
		fmt.Fprintf(&plain, "%s | %s | %s | %s -> %s | %s\n", event.Time.Format(time.RFC3339), event.Kind, event.Record, event.OldIP, event.NewIP, event.Detail)
	}
	for _, part := range []struct{ kind, text string }{{"text/plain", plain.String()}, {"text/html", html.String()}} {
		h := textproto.MIMEHeader{}
		h.Set("Content-Type", part.kind+"; charset=UTF-8")
		h.Set("Content-Transfer-Encoding", "quoted-printable")
		p, err := writer.CreatePart(h)
		if err != nil {
			return nil, err
		}
		q := quotedprintable.NewWriter(p)
		if _, err := q.Write([]byte(part.text)); err != nil {
			return nil, err
		}
		if err := q.Close(); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	var msg bytes.Buffer
	fmt.Fprintf(&msg, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%q\r\n\r\n", from, to, mime.QEncoding.Encode("UTF-8", fmt.Sprintf("[orgmdns] Resumen: %d eventos", len(events))), time.Now().Format(time.RFC1123Z), writer.Boundary())
	msg.Write(body.Bytes())
	return msg.Bytes(), nil
}

func (e *EmailNotifier) SendSummary(events []Event) error {
	if len(events) == 0 {
		return nil
	}
	msg, err := BuildMessage(e.from, e.to, events)
	if err != nil {
		return err
	}
	auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)
	if err := smtp.SendMail(e.smtpHost+":"+e.smtpPort, auth, e.from, []string{e.to}, msg); err != nil {
		return fmt.Errorf("error enviando resumen: %w", err)
	}
	return nil
}
