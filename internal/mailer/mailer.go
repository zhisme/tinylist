package mailer

import (
	"fmt"

	"github.com/zhisme/tinylist/internal/config"
	"gopkg.in/gomail.v2"
)

// Mailer handles email sending
type Mailer struct {
	dialer    *gomail.Dialer
	fromEmail string
	fromName  string
}

// New creates a new Mailer instance
func New(cfg config.SMTPConfig) *Mailer {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.SSL = cfg.TLS && cfg.Port == 465

	return &Mailer{
		dialer:    dialer,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
	}
}

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

// SendCampaign sends a campaign email
func (m *Mailer) SendCampaign(toEmail, toName, subject, textBody, htmlBody, unsubscribeURL string) error {
	// Append unsubscribe link to text body
	textBody = textBody + fmt.Sprintf("\n\n---\nTo unsubscribe, visit: %s", unsubscribeURL)

	// Append unsubscribe link to HTML body if present
	if htmlBody != "" {
		unsubscribeHTML := fmt.Sprintf(`<p style="color: #999; font-size: 12px; margin-top: 40px; border-top: 1px solid #eee; padding-top: 20px;">
<a href="%s" style="color: #999;">Unsubscribe</a></p></body>`, unsubscribeURL)
		htmlBody = htmlBody[:len(htmlBody)-7] + unsubscribeHTML // Replace </body>
	}

	return m.send(toEmail, toName, subject, textBody, htmlBody)
}

// send sends an email
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

// IsConfigured returns true if SMTP is configured
func (m *Mailer) IsConfigured() bool {
	return m.dialer.Host != "" && m.fromEmail != ""
}
