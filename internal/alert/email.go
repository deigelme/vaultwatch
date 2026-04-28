package alert

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailConfig holds SMTP configuration for email notifications.
type EmailConfig struct {
	SMTPHost   string
	SMTPPort   int
	Username   string
	Password   string
	From       string
	To         []string
}

// EmailNotifier sends alert notifications via email.
type EmailNotifier struct {
	cfg EmailConfig
}

// NewEmailNotifier creates a new EmailNotifier with the given configuration.
func NewEmailNotifier(cfg EmailConfig) *EmailNotifier {
	return &EmailNotifier{cfg: cfg}
}

// Send delivers an alert as an email to all configured recipients.
func (e *EmailNotifier) Send(a Alert) error {
	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)

	subject := fmt.Sprintf("[VaultWatch] %s: secret expiring soon", a.Level)
	body := fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s",
		strings.Join(e.cfg.To, ", "),
		e.cfg.From,
		subject,
		a.String(),
	)

	var auth smtp.Auth
	if e.cfg.Username != "" {
		auth = smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.SMTPHost)
	}

	return smtp.SendMail(addr, auth, e.cfg.From, e.cfg.To, []byte(body))
}
