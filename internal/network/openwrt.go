package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

// Function that grabs the openwrt token
func openwrtLogin(openwrtUser, openwrtPw string) (string, error) {
	loginURL := "https://192.168.30.1/cgi-bin/luci/rpc/auth"
	loginBody, err := json.Marshal(map[string]any{
		"id":     1,
		"method": "login",
		"params": []string{openwrtUser, openwrtPw},
	})
	if err != nil {
		return "", fmt.Errorf("encode login body: %w", err)
	}

	r, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(loginBody))
	if err != nil {
		return "", fmt.Errorf("create login request in openwrt: %w", err)
	}
	r.Header.Set("Content-Type", "application/json")

	resp, err := newInsecureClient().Do(r)
	if err != nil {
		return "", fmt.Errorf("send login request in openwrt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login returned status %d in openwrt", resp.StatusCode)
	}

	loginRes := struct {
		Id     int    `json:"id"`
		Result string `json:"result"`
		Error  string `json:"error"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&loginRes); err != nil {
		return "", fmt.Errorf("decode login response in openwrt: %w", err)
	}
	if loginRes.Result == "" {
		return "", fmt.Errorf("empty token in openwrt (wrong credentials?)")
	}

	return loginRes.Result, nil
}

// Function that performs the backup using the token
func openwrtBackup(token string) error {
	backupURL := "https://192.168.30.1/cgi-bin/cgi-backup"
	form := url.Values{}
	form.Set("sessionid", token)

	r, err := http.NewRequest(http.MethodPost, backupURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create backup request in openwrt: %w", err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := newInsecureClient().Do(r)
	if err != nil {
		return fmt.Errorf("send backup request in openwrt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backup returned status %d in openwrt", resp.StatusCode)
	}

	fileName := fmt.Sprintf("/mnt/external/network_backup/backup-Flint2-%s.tar.gz", time.Now().Format("2006-01-02"))

	if err := utils.DownloadFile(fileName, resp.Body); err != nil {
		return err
	}
	log.Printf("[network info] Backup file created in openwrt: %s", fileName)

	// Keeps 2 backups files only
	if err := utils.KeepLastN("/mnt/external/network_backup/backup-Flint2-*.tar.gz", 2); err != nil {
		return err
	}

	return nil
}

func openwrtInit() error {
	openwrtUser := os.Getenv("OPENWRT_USER")
	openwrtPw := os.Getenv("OPENWRT_PW")

	token, err := openwrtLogin(openwrtUser, openwrtPw)
	if err != nil {
		return fmt.Errorf("[network error] openwrt login: %w", err)
	}

	if err := openwrtBackup(token); err != nil {
		return fmt.Errorf("[network error] openwrt backup: %w", err)
	}

	return nil
}
