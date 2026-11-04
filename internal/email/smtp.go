package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/Sahilkumar121/short_url/internal/config"
)

type SMTPSender struct {
	Host     string
	Port     string
	Username string
	Password string
}

func NewSMTPHandler(cfg config.Config) *SMTPSender {
	return &SMTPSender{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
	}
}

func (s *SMTPSender) Send(ctx context.Context, to string, subject string, body string) error {

	// 1. validate email address
	from, err := mail.ParseAddress(s.Username)
	if err != nil {
		return fmt.Errorf("invalid sender username %v", err)
	}

	recipient, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("invalid recipient username %w", err)
	}

	// 2. prevent header injection
	if strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("invalid subject %v", err)
	}

	// 3. connect to smtp
	address := net.JoinHostPort(s.Host, s.Port)

	dailer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	con, err := dailer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("smtp connection failed %w", err)
	}

	defer con.Close()

	// 4. apply deadline
	if deadline, ok := ctx.Deadline(); ok {
		if err := con.SetDeadline(deadline); err != nil {
			return fmt.Errorf("setting deadline failed %w", err)
		}
	}

	stop := context.AfterFunc(ctx, func() {
		con.SetDeadline(time.Now())
	})
	defer stop()

	if err := ctx.Err(); err != nil {
		return err
	}

	// 5. create smtp client
	client, err := smtp.NewClient(con, s.Host)
	if err != nil {
		return fmt.Errorf("creating new client failed %w", err)
	}
	defer client.Close()

	// 6. upgrade connection using STARTTLS
	tlsConfig := &tls.Config{
		ServerName: s.Host,
		MinVersion: tls.VersionTLS12,
	}

	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("smtp STARTTLS failed %w", err)
	}

	// 7. authneticate using SMTP server
	auth := smtp.PlainAuth(
		"",
		s.Username,
		s.Password,
		s.Host,
	)

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp authenticate failed %w", err)
	}

	// 8. specify the sender
	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("smtp sender failed %w", err)
	}

	// 9. specify the receiver
	if err := client.Rcpt(recipient.Address); err != nil {
		return fmt.Errorf("smtp recipient failed %w", err)
	}

	message := []byte(
		"From: " + from.String() + "\r\n" +
			"To: " + recipient.String() + "\r\n" +
			"Subject: " + mime.QEncoding.Encode("UTF-8", subject) + "\r\n" +
			body + "\r\n",
	)

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data failed %w", err)
	}

	if _, err := writer.Write(message); err != nil {
		if err := writer.Close(); err != nil {
			return fmt.Errorf("writer close failed %w", err)
		}
		return fmt.Errorf("write email failed %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("client quit failed %w", err)
	}

	return nil
}
