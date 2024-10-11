package email

import (
	"bytes"
	"fmt"
	"log"
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
	// FIXME: when supervisorctl is correctly implemented
	// go to /var/log/gbackup/gbackup-today_date.err.log and read
	// this file instead of /root/gbackup/logstdout
	// then send this file to the email
	logstdoutFile, err := os.ReadFile("/root/gbackup/logstdout")
	if err != nil {
		return "", fmt.Errorf("[email error] could not read file logstdout: %s\n", err)
	}

	buf := bytes.NewBuffer(logstdoutFile)

	for _, s := range strings.Split(buf.String(), "\n") {
		if strings.Contains(s, "Starting") || strings.Contains(s, "Executing") {
			e.Content += fmt.Sprintf("%s<br><br>", s)
		}
	}

	templ, err := template.New("email_template.html").ParseFiles("/root/gbackup/email_template.html")
	if err != nil {
		return "", fmt.Errorf("[email error] could not parse files: %s\n", err)
	}

	var outTemp bytes.Buffer
	if err := templ.Execute(&outTemp, e); err != nil {
		return "", fmt.Errorf("[email error] could not execute template: %s\n", err)
	}

	return outTemp.String(), nil
}

func SendEmail(body *EmailTemplate) error {
	senderEmail, senderPass := getEnvs()

	s := &Smtp{}
	s.setSmtpValues()

	recipientEmail := "brunoalexandre3@hotmail.com"
	headers := "Content-Type: text/html; charset=ISO-8859-1\r\n" // used to send HTML

	from := fmt.Sprintf("From: <%s>\r\n", senderEmail)
	to := fmt.Sprintf("To: <%s>\r\n", recipientEmail)
	subject := "Subject: Backup executed\r\n"

	log.Printf("[email] building email body\n")
	finalBody, err := buildEmail(body)
	if err != nil {
		return fmt.Errorf("[email error] could not build email template: %s", err)
	}

	msg := headers + from + to + subject + "\r\n" + finalBody + "\r\n"

	auth := smtp.PlainAuth("", senderEmail, senderPass, s.Host)

	if err := smtp.SendMail(s.Address, auth, senderEmail, []string{recipientEmail}, []byte(msg)); err != nil {
		return fmt.Errorf("[email error] could not send email: %s", err)
	}

	log.Printf("[email info] sent email to %s\n", recipientEmail)

	return nil
}
