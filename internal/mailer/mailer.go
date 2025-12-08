package mailer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zhisme/tinylist/internal/config"
	"gopkg.in/gomail.v2"
)

// Default timeout for SMTP operations
const defaultSendTimeout = 30 * time.Second

// Mailer handles email sending
type Mailer struct {
	dialer      *gomail.Dialer
	fromEmail   string
	fromName    string
	sendTimeout time.Duration
	host        string
	port        int
	username    string
	password    string
	tls         bool
}

// New creates a new Mailer instance
func New(cfg config.SMTPConfig) *Mailer {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.SSL = cfg.TLS && cfg.Port == 465

	return &Mailer{
		dialer:      dialer,
		fromEmail:   cfg.FromEmail,
		fromName:    cfg.FromName,
		sendTimeout: defaultSendTimeout,
		host:        cfg.Host,
		port:        cfg.Port,
		username:    cfg.Username,
		password:    cfg.Password,
		tls:         cfg.TLS,
	}
}

// Reconfigure updates the mailer with new SMTP settings
func (m *Mailer) Reconfigure(host string, port int, username, password, fromEmail, fromName string, tls bool) {
	dialer := gomail.NewDialer(host, port, username, password)
	dialer.SSL = tls && port == 465

	m.dialer = dialer
	m.fromEmail = fromEmail
	m.fromName = fromName
	m.host = host
	m.port = port
	m.username = username
	m.password = password
	m.tls = tls
}

// SendTest sends a test email to verify SMTP configuration
func (m *Mailer) SendTest(toEmail string) error {
	subject := "TinyList - Test Email"
	textBody := fmt.Sprintf(`This is a test email from TinyList.

If you received this email, your SMTP configuration is working correctly.

Best regards,
%s`, m.fromName)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
<h2>TinyList - Test Email</h2>
<p>This is a test email from TinyList.</p>
<p>If you received this email, your SMTP configuration is working correctly.</p>
<p style="margin-top: 40px;">Best regards,<br>%s</p>
</body>
</html>`, m.fromName)

	return m.send(toEmail, "", subject, textBody, htmlBody)
}

// TODO: move to separate email template files if they get more complex
// SendVerification sends a verification email
func (m *Mailer) SendVerification(toEmail, toName, verifyURL string) error {
	subject := "Please verify your email address"
	textBody := fmt.Sprintf(`Hi %s,

Thanks for subscribing! Please verify your email address by clicking the link below:

%s

If you didn't subscribe to this list, you can safely ignore this email.

Best regards,
%s`, toName, verifyURL, m.fromName)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
<h2>Verify your email address</h2>
<p>Hi %s,</p>
<p>Thanks for subscribing! Please verify your email address by clicking the button below:</p>
<p style="margin: 30px 0;">
  <a href="%s" style="background-color: #4CAF50; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">Verify Email</a>
</p>
<p>Or copy and paste this link into your browser:</p>
<p style="word-break: break-all; color: #666;">%s</p>
<p style="color: #999; font-size: 12px; margin-top: 40px;">
If you didn't subscribe to this list, you can safely ignore this email.
</p>
</body>
</html>`, toName, verifyURL, verifyURL)

	return m.send(toEmail, toName, subject, textBody, htmlBody)
}

// SendCampaign sends a campaign email with context support for cancellation/timeout
func (m *Mailer) SendCampaign(ctx context.Context, toEmail, toName, subject, textBody, htmlBody, unsubscribeURL string) error {
	// Append unsubscribe link to text body
	textBody = textBody + fmt.Sprintf("\n\n---\nTo unsubscribe, visit: %s", unsubscribeURL)

	// Append unsubscribe link to HTML body if present
	if htmlBody != "" {
		unsubscribeHTML := fmt.Sprintf(`<p style="color: #999; font-size: 12px; margin-top: 40px; border-top: 1px solid #eee; padding-top: 20px;">
<a href="%s" style="color: #999;">Unsubscribe</a></p>`, unsubscribeURL)
		// Try to insert before </body>, otherwise just append
		if idx := strings.LastIndex(strings.ToLower(htmlBody), "</body>"); idx != -1 {
			htmlBody = htmlBody[:idx] + unsubscribeHTML + htmlBody[idx:]
		} else {
			htmlBody = htmlBody + unsubscribeHTML
		}
	}

	return m.sendWithContext(ctx, toEmail, toName, subject, textBody, htmlBody)
}

// send sends an email (blocking, no timeout)
func (m *Mailer) send(toEmail, toName, subject, textBody, htmlBody string) error {
	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", m.fromEmail, m.fromName)
	msg.SetAddressHeader("To", toEmail, toName)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", textBody)
	if htmlBody != "" {
		msg.AddAlternative("text/html", htmlBody)
	}

	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// sendWithContext sends an email with context support for cancellation/timeout
func (m *Mailer) sendWithContext(ctx context.Context, toEmail, toName, subject, textBody, htmlBody string) error {
	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", m.fromEmail, m.fromName)
	msg.SetAddressHeader("To", toEmail, toName)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", textBody)
	if htmlBody != "" {
		msg.AddAlternative("text/html", htmlBody)
	}

	// Create a timeout context if parent doesn't have deadline
	sendCtx := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		sendCtx, cancel = context.WithTimeout(ctx, m.sendTimeout)
		defer cancel()
	}

	// Run send in goroutine so we can respect context cancellation
	errCh := make(chan error, 1)
	go func() {
		errCh <- m.dialer.DialAndSend(msg)
	}()

	select {
	case <-sendCtx.Done():
		return fmt.Errorf("send cancelled or timed out: %w", sendCtx.Err())
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
		return nil
	}
}

// IsConfigured returns true if SMTP is configured
func (m *Mailer) IsConfigured() bool {
	return m.host != "" && m.fromEmail != ""
}
