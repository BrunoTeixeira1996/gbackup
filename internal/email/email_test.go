package email_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/email"
	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

func TestSendEmail(t *testing.T) {
	b := []targets.BackupResult{
		{
			TargetName: "Target1",
			ElapsedTime: utils.ElapsedTime{
				Target: "Target1",
				Value:  1234,
			},
			TargetSize: utils.TargetSize{
				Name:   "Target1",
				Before: 22222,
				After:  33333,
			},
			Err: nil,
		},
		{
			TargetName:  "Target2",
			ElapsedTime: utils.ElapsedTime{},
			TargetSize: utils.TargetSize{
				Name:   "Target2",
				Before: 22222,
				After:  33333,
			},
			Err: fmt.Errorf("some error"),
		},
		{
			TargetName: "Target3",
			ElapsedTime: utils.ElapsedTime{
				Target: "Target3",
				Value:  12345,
			},
			TargetSize: utils.TargetSize{
				Name:   "Target3",
				Before: 22222,
				After:  33333,
			},
			Err: nil,
		},
		{
			TargetName: "Target4",
			ElapsedTime: utils.ElapsedTime{
				Target: "Target4",
				Value:  123456,
			},
			TargetSize: utils.TargetSize{
				Name:   "Target4",
				Before: 22222,
				After:  33333,
			},
			Err: nil,
		},
	}

	e := email.EmailClient{}
	e.InitEmailClient()

	err := e.SendEmail(b, "/home/brun0/Desktop/personal/gbackup/internal/email/testlog.txt")
	if err != nil {
		log.Println(err)
		t.Errorf("failed to send email: %v", err)
	}
}
