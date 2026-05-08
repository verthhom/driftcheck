package notify

import "net/smtp"

// SetDialFunc replaces the underlying SMTP dial function.
// Intended for use in tests only.
func (n *EmailNotifier) SetDialFunc(
	fn func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error,
) {
	n.dial = fn
}
