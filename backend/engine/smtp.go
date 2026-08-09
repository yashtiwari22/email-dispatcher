package engine

import (
	"fmt"
	"net/smtp"
)

// SMTPSender defines the interface for sending emails.
type SMTPSender interface {
	SendEmail(to string, subject string, body string) error
}

// RealSMTPSender implements SMTPSender using net/smtp.
type RealSMTPSender struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// NewSMTPSender initializes a RealSMTPSender instance.
func NewSMTPSender(host string, port int, username, password, from string) *RealSMTPSender {
	return &RealSMTPSender{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

// SendEmail dispatches an email via net/smtp.
func (s *RealSMTPSender) SendEmail(to string, subject string, body string) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	var auth smtp.Auth
	if s.Username != "" && s.Password != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
		"%s", to, s.From, subject, body))

	if err := smtp.SendMail(addr, auth, s.From, []string{to}, msg); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	return nil
}

// MockSMTPSender is useful for unit testing without a running SMTP server.
type MockSMTPSender struct {
	SentEmails []struct {
		To      string
		Subject string
		Body    string
	}
	ShouldFail bool
}

func (m *MockSMTPSender) SendEmail(to string, subject string, body string) error {
	if m.ShouldFail {
		return fmt.Errorf("mock smtp error: failed to send email to %s", to)
	}
	m.SentEmails = append(m.SentEmails, struct {
		To      string
		Subject string
		Body    string
	}{To: to, Subject: subject, Body: body})
	return nil
}
