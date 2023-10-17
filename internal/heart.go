package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

func getAssociatedIpFromMAC(mac string) (string, error) {
	command := fmt.Sprintf("ip neighbor | grep '%s'", mac)
	out, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		return "", fmt.Errorf("Could not grep that mac address: %v", err)
	}
	return strings.Split(string(out), " ")[0], nil
}

func IsAlive(mac string) (bool, error) {
	targetIP, err := getAssociatedIpFromMAC(mac)
	if err != nil {
		return false, err
	}
	out, err := exec.Command("ping", targetIP, "-c 2").Output()
	if err != nil {
		return false, fmt.Errorf("Could not ping that IP: %v", err)
	}

	if strings.Contains(string(out), "Destination Host Unreachable") {
		return false, nil
	} else {
		return true, nil
	}
}
