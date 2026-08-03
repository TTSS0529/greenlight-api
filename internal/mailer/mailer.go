package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

type smtpSender interface {
	DialAndSend(...*mail.Message) error
}

type Mailer struct {
	dialer     smtpSender
	sender     string
	sleep      func(time.Duration)
	newMessage func() *mail.Message
}

func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second
	return Mailer{
		dialer: dialer,
		sender: sender,
		sleep:  time.Sleep,
		newMessage: func() *mail.Message {
			return mail.NewMessage()
		},
	}
}

func (m Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").Option("missingkey=error").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}
	plainbody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainbody, "plainBody", data)
	if err != nil {
		return err
	}
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}
	msg := m.newMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainbody.String())
	msg.AddAlternative("text/html", htmlBody.String())
	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if nil == err {
			return nil
		}
		m.sleep(500 * time.Millisecond)
	}
	return err
}
