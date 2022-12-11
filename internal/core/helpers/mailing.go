package helpers

import (
	"bytes"
	"fmt"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"html/template"
	"net/smtp"
	"os"
	"strings"
)

func ParseTemplate(templateFileName string, data interface{}) (*string, error) {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return nil, err
	}
	str := buf.String()
	return &str, nil
}

func SendEmail(email string, subject string, template string, templateData interface{}) error {

	sender := os.Getenv("MAIL_FROM")

	to := []string{
		email,
	}

	user := os.Getenv("SMTP_LOGIN")
	if user == "" {
		return errors.New("Mailserver not found")
	}
	password := os.Getenv("SMTP_PASSWORD")

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	auth := smtp.PlainAuth("", user, password, host)

	body, err := ParseTemplate(template, templateData)

	if err != nil {
		return err
	}

	r := Mail{
		Sender:  sender,
		To:      to,
		Subject: subject,
		Body:    *body,
	}

	msg := []byte(r.BuildMessage())

	err = smtp.SendMail(addr, auth, sender, to, msg)

	if err != nil {
		return err
	}
	return nil
}
func (m *Mail) BuildMessage() string {
	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", m.Sender)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(m.To, ";"))
	msg += fmt.Sprintf("Subject: %s\r\n", m.Subject)
	msg += fmt.Sprintf("\r\n%s\r\n", m.Body)

	return msg
}

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}
