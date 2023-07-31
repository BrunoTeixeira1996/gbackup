package internal

import (
	"fmt"
	"net/smtp"
	"os"
)

type Smtp struct {
	Host    string
	Port    string
	Address string
}

func (s *Smtp) setSmtpValues() {
	s.Host = "smtp.gmail.com"
	s.Port = "587"
	s.Address = s.Host + ":" + s.Port
}

func getEnvs() (string, string) {
	senderEmail := os.Getenv("SENDEREMAIL")
	senderPass := os.Getenv("SENDERPASS")

	return senderEmail, senderPass
}

func SendEmail(body string) error {
	senderEmail, senderPass := getEnvs()
	if senderEmail == "" || senderPass == "" {
		return fmt.Errorf("Please setup the SENDEREMAIL and SENDERPASS env vars")
	}

	s := &Smtp{}
	s.setSmtpValues()

	recipientEmail := "brunoalexandre3@hotmail.com"

	from := fmt.Sprintf("From: <%s>\r\n", senderEmail)
	to := fmt.Sprintf("To: <%s>\r\n", recipientEmail)
	subject := "Subject: Backup executed\r\n"
	msg := from + to + subject + "\r\n" + body + "\r\n"

	auth := smtp.PlainAuth("", senderEmail, senderPass, s.Host)

	if err := smtp.SendMail(s.Address, auth, senderEmail, []string{recipientEmail}, []byte(msg)); err != nil {
		return err
	}

	return nil
}
