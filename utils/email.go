package utils

import (
	"fmt"
	"net/smtp"

	"event-messenger.com/config"
)

func SendEmailNotification(toEmail, subject, htmlContent string) error {
	auth := smtp.PlainAuth("", config.App.SMTPUsername, config.App.SMTPPassword, config.App.SMTPServer)

	// Compose email headers and body
	headers := make(map[string]string)
	headers["From"] = config.App.FromEmail
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlContent

	// Send email
	addr := fmt.Sprintf("%s:%d", config.App.SMTPServer, config.App.SMTPPort)
	err := smtp.SendMail(addr, auth, config.App.FromEmail, []string{toEmail}, []byte(message))

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	fmt.Printf("Email sent successfully to: %s\n", toEmail)
	return nil
}
