package notify

import (
	"fmt"
	"net/smtp"
	"strings"

	"driftcheck/internal/alert"
)

// EmailConfig holds the configuration for sending email notifications.
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	From     string
	To       []string
	Username string
	Password string
}

// EmailNotifier sends drift alerts via email.
type EmailNotifier struct {
	cfg  EmailConfig
	dial func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error
}

// NewEmailNotifier creates an EmailNotifier with the given configuration.
func NewEmailNotifier(cfg EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		cfg:  cfg,
		dial: smtp.SendMail,
	}
}

// Notify sends an email alert for the given drift result.
func (n *EmailNotifier) Notify(result alert.Result) error {
	if len(n.cfg.To) == 0 {
		return fmt.Errorf("email notifier: no recipients configured")
	}
	if n.cfg.SMTPHost == "" {
		return fmt.Errorf("email notifier: SMTP host is required")
	}

	subject := fmt.Sprintf("[driftcheck] %s drift detected (severity: %s)",
		result.ServiceName, result.Severity)

	var body strings.Builder
	body.WriteString(fmt.Sprintf("Service: %s\n", result.ServiceName))
	body.WriteString(fmt.Sprintf("Severity: %s\n", result.Severity))
	body.WriteString(fmt.Sprintf("Drifted keys (%d):\n", len(result.DriftedKeys)))
	for _, k := range result.DriftedKeys {
		body.WriteString(fmt.Sprintf("  - %s\n", k))
	}

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		n.cfg.From,
		strings.Join(n.cfg.To, ", "),
		subject,
		body.String(),
	))

	addr := fmt.Sprintf("%s:%d", n.cfg.SMTPHost, n.cfg.SMTPPort)
	var auth smtp.Auth
	if n.cfg.Username != "" {
		auth = smtp.PlainAuth("", n.cfg.Username, n.cfg.Password, n.cfg.SMTPHost)
	}

	if err := n.dial(addr, auth, n.cfg.From, n.cfg.To, msg); err != nil {
		return fmt.Errorf("email notifier: send failed: %w", err)
	}
	return nil
}
