package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// SMTPConfig holds SMTP connection parameters.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	TLS      bool // true = implicit TLS (port 465); false = STARTTLS or plain
	Insecure bool // skip TLS cert verification
}

type smtpMailer struct {
	cfg SMTPConfig
}

// NewSMTP constructs a real SMTP Mailer.
// If cfg.Host is empty, returns nil (caller should fall back to stub).
func NewSMTP(cfg SMTPConfig) Mailer {
	if cfg.Host == "" {
		return nil
	}
	if cfg.Port == 0 {
		if cfg.TLS {
			cfg.Port = 465
		} else {
			cfg.Port = 587
		}
	}
	return &smtpMailer{cfg: cfg}
}

func (m *smtpMailer) Send(_ context.Context, msg Message) error {
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	raw := m.buildRaw(msg)

	if m.cfg.TLS {
		return m.sendTLS(addr, raw, msg.To)
	}
	return m.sendSTARTTLS(addr, raw, msg.To)
}

func (m *smtpMailer) sendTLS(addr string, raw []byte, to string) error {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: m.cfg.Insecure, //nolint:gosec
		ServerName:         m.cfg.Host,
	}
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 15 * time.Second}, "tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}
	c, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer func() { _ = c.Close() }()
	return m.doSend(c, raw, to)
}

func (m *smtpMailer) sendSTARTTLS(addr string, raw []byte, to string) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer func() { _ = c.Close() }()

	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: m.cfg.Insecure, //nolint:gosec
			ServerName:         m.cfg.Host,
		}
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}
	return m.doSend(c, raw, to)
}

func (m *smtpMailer) doSend(c *smtp.Client, raw []byte, to string) error {
	if m.cfg.Username != "" {
		auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	fromAddr := extractAddr(m.cfg.From)
	if err := c.Mail(fromAddr); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT TO: %w", err)
	}
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}
	if _, err = wc.Write(raw); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	return wc.Close()
}

func (m *smtpMailer) buildRaw(msg Message) []byte {
	var buf bytes.Buffer
	from := m.cfg.From
	if from == "" {
		from = m.cfg.Username
	}
	subj := mime.QEncoding.Encode("UTF-8", msg.Subject)
	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + msg.To + "\r\n")
	buf.WriteString("Subject: " + subj + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")

	if msg.BodyHTML != "" {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(msg.BodyHTML)
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(msg.Body)
	}
	return buf.Bytes()
}

// extractAddr pulls the bare address from "Name <addr>" or returns the string as-is.
func extractAddr(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "<"); i >= 0 {
		if j := strings.Index(s[i:], ">"); j >= 0 {
			return strings.TrimSpace(s[i+1 : i+j])
		}
	}
	return s
}
