package notify

import (
	"fmt"
	"net/smtp"
	"time"
)

type EmailNotifier struct {
	from     string
	to       string
	password string
	smtpHost string
	smtpPort string
}

func NewEmailNotifier(from, to, password, smtpHost, smtpPort string) *EmailNotifier {
	return &EmailNotifier{
		from:     from,
		to:       to,
		password: password,
		smtpHost: smtpHost,
		smtpPort: smtpPort,
	}
}

// SendDNSUpdateNotification envía un correo notificando el cambio de IP en un registro DNS
func (e *EmailNotifier) SendDNSUpdateNotification(recordName, oldIP, newIP string) error {
	subject := fmt.Sprintf("[orgmdns] DNS actualizado: %s", recordName)
	body := fmt.Sprintf(`Hola,

El registro DNS ha sido actualizado automáticamente por orgmdns.

Detalles:
- Registro: %s
- IP anterior: %s
- IP nueva: %s
- Fecha/hora: %s

Este es un mensaje automático, por favor no respondas.

--
orgmdns
`, recordName, oldIP, newIP, time.Now().Format("2006-01-02 15:04:05 MST"))

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", e.from, e.to, subject, body)

	auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)

	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)
	err := smtp.SendMail(addr, auth, e.from, []string{e.to}, []byte(message))
	if err != nil {
		return fmt.Errorf("error enviando correo: %w", err)
	}

	return nil
}

// SendStartupNotification envía un correo cuando inicia el verificador DNS
func (e *EmailNotifier) SendStartupNotification(currentIP string, recordNames []string) error {
	subject := "[orgmdns] Verificador DNS corriendo"
	
	// Formatear lista de subdominios
	recordsList := ""
	for i, name := range recordNames {
		recordsList += fmt.Sprintf("- %s", name)
		if i < len(recordNames)-1 {
			recordsList += "\n"
		}
	}
	
	body := fmt.Sprintf(`Hola,

El verificador DNS ha iniciado correctamente.

Detalles:
- IP pública detectada: %s
- Fecha/hora de inicio: %s
- Subdominios configurados:
%s

El sistema está monitoreando los registros DNS configurados.

Este es un mensaje automático, por favor no respondas.

--
orgmdns
`, currentIP, time.Now().Format("2006-01-02 15:04:05 MST"), recordsList)

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", e.from, e.to, subject, body)

	auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)

	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)
	err := smtp.SendMail(addr, auth, e.from, []string{e.to}, []byte(message))
	if err != nil {
		return fmt.Errorf("error enviando correo de inicio: %w", err)
	}

	return nil
}

// SendConnectionRestoredNotification envía un correo cuando se restaura la conexión
func (e *EmailNotifier) SendConnectionRestoredNotification(duration time.Duration) error {
	subject := "[orgmdns] Conexión a internet restaurada"
	
	// Formatear duración de forma legible
	var durationStr string
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		durationStr = fmt.Sprintf("%d día(s), %d hora(s), %d minuto(s)", days, hours, minutes)
	} else if hours > 0 {
		durationStr = fmt.Sprintf("%d hora(s), %d minuto(s)", hours, minutes)
	} else {
		durationStr = fmt.Sprintf("%d minuto(s)", minutes)
	}

	body := fmt.Sprintf(`Hola,

La conexión a internet se ha restablecido en el sistema DNS.

Detalles:
- Tiempo sin conexión: %s
- Fecha/hora de restauración: %s

El verificador DNS ha reanudado su funcionamiento normal.

Este es un mensaje automático, por favor no respondas.

--
orgmdns
`, durationStr, time.Now().Format("2006-01-02 15:04:05 MST"))

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", e.from, e.to, subject, body)

	auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)

	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)
	err := smtp.SendMail(addr, auth, e.from, []string{e.to}, []byte(message))
	if err != nil {
		return fmt.Errorf("error enviando correo de restauración: %w", err)
	}

	return nil
}

// SendErrorNotification envía un correo notificando un error (opcional, para uso futuro)
func (e *EmailNotifier) SendErrorNotification(errorMsg string) error {
	subject := "[orgmdns] Error en la aplicación"
	body := fmt.Sprintf(`Hola,

Se ha producido un error en orgmdns:

%s

Fecha/hora: %s

--
orgmdns
`, errorMsg, time.Now().Format("2006-01-02 15:04:05 MST"))

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", e.from, e.to, subject, body)

	auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)

	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)
	err := smtp.SendMail(addr, auth, e.from, []string{e.to}, []byte(message))
	if err != nil {
		return fmt.Errorf("error enviando correo de error: %w", err)
	}

	return nil
}
