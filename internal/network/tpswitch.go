package network

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

// Function that grabs the tpswitch token
func tpswitchLogin(tpswitchUser, tpswitchPw string) (string, error) {
	loginURL := "http://tl-switch.lan/logon.cgi"
	form := url.Values{
		"username":  {tpswitchUser},
		"password":  {tpswitchPw},
		"cpassword": {""},
		"logon":     {"Login"},
	}

	r, err := http.NewRequest(http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("create login request in tpswitch: %w", err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := newInsecureClient().Do(r)
	if err != nil {
		return "", fmt.Errorf("send login request in tpswitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login returned status %d in tpswitch", resp.StatusCode)
	}

	for _, c := range resp.Cookies() {
		if c.Name == "H_P_SSID" {
			return c.Value, nil
		}
	}
	return "", fmt.Errorf("H_P_SSID cookie not found in tpswitch login")
}

// Function that performs the backup using the token
func tpswitchBackup(token string) error {
	backupURL := "http://tl-switch.lan/config_back.cgi?btnBackup=Backup+Config"

	r, err := http.NewRequest(http.MethodGet, backupURL, nil)
	if err != nil {
		return fmt.Errorf("create backup request in tpswitch: %w", err)
	}
	r.AddCookie(&http.Cookie{Name: "H_P_SSID", Value: token})

	resp, err := newInsecureClient().Do(r)
	if err != nil {
		return fmt.Errorf("send backup request in tpswitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backup returned status %d in tpswitch", resp.StatusCode)
	}

	fileName := fmt.Sprintf("/mnt/external/network_backup/backup-Tpswitch-%s.cfg", time.Now().Format("2006-01-02"))

	if err := utils.DownloadFile(fileName, resp.Body); err != nil {
		return err
	}
	log.Printf("[network info] Backup file created in tpswitch: %s", fileName)

	// Keeps 2 backups files only
	if err := utils.KeepLastN("/mnt/external/network_backup/backup-Tpswitch-*.cfg", 2); err != nil {
		return err
	}

	return nil
}

func tpswitchInit() error {
	tpswitchUser := os.Getenv("TPSWITCH_USER")
	tpswitchPw := os.Getenv("TPSWITCH_PW")

	token, err := tpswitchLogin(tpswitchUser, tpswitchPw)
	if err != nil {
		return fmt.Errorf("[network error] tpswitch login: %w", err)
	}

	if err := tpswitchBackup(token); err != nil {
		return fmt.Errorf("[network error] tpswitch backup: %w", err)
	}

	return nil
}
