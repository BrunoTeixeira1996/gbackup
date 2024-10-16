package email

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
)

type EmailClient struct {
	Email         string
	Password      string
	Smtp          Smtp
	EmailTemplate EmailTemplate
}

type EmailTemplate struct {
	Timestamp     string
	TotalBackups  int
	BackupResults []targets.BackupResult
	Content       []string
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

// Find current date and grab the logs for that date
// it slices the logs around ================ string
func (eT *EmailTemplate) extractLogForTheDay(logFilePath string) error {
	var (
		capturing    bool
		date         = time.Now().Format("2006/01/02")
		startingDate = fmt.Sprintf("Starting Gbackup: %s", date)
		endingDate   = fmt.Sprintf("Ending Gbackup: %s", date)
		tempContent  []string
	)

	logFile, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("[email error] could not read file logstdout: %s\n", err)
	}
	defer logFile.Close()

	scanner := bufio.NewScanner(logFile)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, startingDate) {
			capturing = true
			tempContent = append(tempContent, "==========================================================================")
			tempContent = append(tempContent, line)
			continue
		}

		if strings.Contains(line, endingDate) {
			if capturing {
				tempContent = append(tempContent, line)
				tempContent = append(tempContent, "==========================================================================")
				capturing = false
				break
			}
		}

		if capturing {
			tempContent = append(tempContent, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("[extract log for the day error] had errors in scanner: %s\n", err)
	}

	eT.Content = tempContent

	return nil
}

func (e *EmailClient) InitEmailClient() {
	e.Email = os.Getenv("SENDEREMAIL")
	e.Password = os.Getenv("SENDERPASS")

	s := Smtp{}
	s.setSmtpValues()
	e.Smtp = s
}

// Read backupResults and add that to html template
// Read /var/log/gbackup/gbackup.err.log, grab only the specific backup
// and add that to html template
func (e *EmailClient) buildEmail(backupResults []targets.BackupResult, logPathFile string) (string, error) {
	tempEmailTemplate := EmailTemplate{
		Timestamp:     time.Now().String(),
		TotalBackups:  len(backupResults),
		BackupResults: backupResults,
	}

	if err := tempEmailTemplate.extractLogForTheDay(logPathFile); err != nil {
		return "", err
	}

	e.EmailTemplate = tempEmailTemplate

	// prod
	templ, err := template.New("email.html").ParseFiles("/home/brun0/src/gbackup/email.html")
	// debug
	//templ, err := template.New("email.html").ParseFiles("/home/brun0/Desktop/personal/gbackup/internal/email/email.html")

	if err != nil {
		return "", fmt.Errorf("[email error] could not parse email html template: %s\n", err)
	}

	var outTemp bytes.Buffer
	if err := templ.Execute(&outTemp, e.EmailTemplate); err != nil {
		return "", fmt.Errorf("[email error] could not execute template: %s\n", err)
	}

	return outTemp.String(), nil
}

func (e *EmailClient) SendEmail(backupResults []targets.BackupResult, logPathFile string) error {
	recipientEmail := "brunoalexandre3@hotmail.com"
	headers := "Content-Type: text/html; charset=ISO-8859-1\r\n" // used to send HTML

	from := fmt.Sprintf("From: <%s>\r\n", e.Email)
	to := fmt.Sprintf("To: <%s>\r\n", recipientEmail)
	subject := "Subject: Backup executed\r\n"

	log.Printf("[email] building email body\n")

	finalBody, err := e.buildEmail(backupResults, logPathFile)

	if err != nil {
		return err
	}

	msg := headers + from + to + subject + "\r\n" + finalBody + "\r\n"

	auth := smtp.PlainAuth("", e.Email, e.Password, e.Smtp.Host)

	if err := smtp.SendMail(e.Smtp.Address, auth, e.Email, []string{recipientEmail}, []byte(msg)); err != nil {
		return fmt.Errorf("[email error] could not send email: %s", err)
	}

	log.Printf("[email info] sent email to %s\n", recipientEmail)

	return nil
}
