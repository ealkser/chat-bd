package services

import (
	"log"
	"net/smtp"
)

// EmailService предоставляет методы для отправки писем
type EmailService struct {
	from     string
	password string
	smtpHost string
	smtpPort string
}

// NewEmailService создаёт новый сервис для отправки писем
func NewEmailService(from, password, host, port string) *EmailService {
	return &EmailService{
		from:     from,
		password: password,
		smtpHost: host,
		smtpPort: port,
	}
}

// SendVerificationCode отправляет 6-значный код на email
func (e *EmailService) SendVerificationCode(toEmail, name, code string) error {
	subject := "Код подтверждения"
	body := generateVerificationEmailBody(name, code)

	msg := []byte(
		"To: " + toEmail + "\r\n" +
			"From: " + e.from + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)
	err := smtp.SendMail(e.smtpHost+":"+e.smtpPort, auth, e.from, []string{toEmail}, msg)
	if err != nil {
		return err
	}

	log.Printf("Код подтверждения отправлен на %s", toEmail)
	return nil
}

// generateVerificationEmailBody формирует тело письма
func generateVerificationEmailBody(name, code string) string {
	return `
Здравствуйте, ` + name + `!

Ваш код подтверждения: **` + code + `**

Он действителен в течение 10 минут.

Если вы не регистрировались — проигнорируйте это письмо.

С уважением,
Команда приложения`
}
