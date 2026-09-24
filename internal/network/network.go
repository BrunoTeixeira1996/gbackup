package network

import (
	"crypto/tls"
	"net/http"
)

// The router and switch use a self-signed certificate
func newInsecureClient() *http.Client {
	return &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
}

// Main function to perform the network backup
func ExecuteNetworkBackup() error {
	if err := openwrtInit(); err != nil {
		return err
	}
	if err := tpswitchInit(); err != nil {
		return err
	}
	return nil
}
