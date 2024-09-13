package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

type EmailTemplate struct {
	Timestamp          string
	Totalbackups       int
	Totalbackupsuccess int
	PiTemp             string
	Content            string
	ElapsedTimes       []utils.ElapsedTime
	TotalElapsedTime   float64
	TargetsSize        []utils.TargetSize
}

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

func buildEmail(e *EmailTemplate) (string, error) {
	logstdoutFile, err := os.ReadFile("/root/gbackup/logstdout")
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(logstdoutFile)

	for _, s := range strings.Split(buf.String(), "\n") {
		if strings.Contains(s, "Starting") || strings.Contains(s, "Executing") {
			e.Content += fmt.Sprintf("%s<br><br>", s)
		}
	}

	templ, err := template.New("email_template.html").ParseFiles("/root/gbackup/email_template.html")
	if err != nil {
		return "", err
	}

	var outTemp bytes.Buffer
	if err := templ.Execute(&outTemp, e); err != nil {
		return "", err
	}

	return outTemp.String(), nil
}

func SendEmail(body *EmailTemplate) error {
	senderEmail, senderPass := getEnvs()
	if senderEmail == "" || senderPass == "" {
		return fmt.Errorf("Please setup the SENDEREMAIL and SENDERPASS env vars")
	}

	s := &Smtp{}
	s.setSmtpValues()

	recipientEmail := "brunoalexandre3@hotmail.com"
	headers := "Content-Type: text/html; charset=ISO-8859-1\r\n" // used to send HTML

	from := fmt.Sprintf("From: <%s>\r\n", senderEmail)
	to := fmt.Sprintf("To: <%s>\r\n", recipientEmail)
	subject := "Subject: Backup executed\r\n"

	finalBody, err := buildEmail(body)
	if err != nil {
		return fmt.Errorf("Error while building email template: %w", err)
	}

	msg := headers + from + to + subject + "\r\n" + finalBody + "\r\n"

	auth := smtp.PlainAuth("", senderEmail, senderPass, s.Host)

	if err := smtp.SendMail(s.Address, auth, senderEmail, []string{recipientEmail}, []byte(msg)); err != nil {
		return err
	}

	return nil
}
