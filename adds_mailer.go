package gomixer

import (
	"crypto/tls"
	"fmt"
	gml "gopkg.in/gomail.v2"
	"errors"
)

//---------------------------------------------------------------------------
//  MAIL: поддержка почты в рамках сессии для обработчиков запросов
//---------------------------------------------------------------------------
type Mail struct {
	MailMessage MailMessage
}
type MailMessage struct {
	To         string
	From       string
	Message    string
	Subject    string
	FileAttach string `fullpath to file`
	Host       string
	Port       int
	Username   string
	Password   string
}

//---------------------------------------------------------------------------
//  MAIL
//---------------------------------------------------------------------------
func NewMail() *Mail {
	return &Mail{MailMessage:MailMessage{}}
}
func (m *Mail) NewMailMessage(to, from, message, subject, host, username, password, fileattach string, port int) MailMessage {
	return MailMessage{
		To:to,
		From: from,
		Message: message,
		Subject: subject,
		Host: host,
		Port: port,
		Username: username,
		Password: password,
		FileAttach: fileattach,
	}
}
func (mail MailMessage) SendMail() (error) {
	ds := gml.NewDialer(mail.Host, mail.Port, mail.Username, mail.Password)
	ds.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	m := gml.NewMessage()
	m.SetHeader("From", mail.From)
	m.SetHeader("To", mail.To)
	//	m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", mail.Subject)
	m.SetBody("text/html", mail.Message)
	if mail.FileAttach != "" {
		m.Attach(mail.FileAttach)
	}
	if err := ds.DialAndSend(m); err != nil {
		fmt.Printf("[sendemail] ошибка отправки сообщения %v\n", err)
		return errors.New(fmt.Sprintf("[sendemail] ошибка отправки сообщения %v\n", err))
	}
	return nil
}