package utils

import (
	"log"
	"os"
	"strings"
)

func checkEnvVars() bool {
	envVars := map[string]string{
		"SENDEREMAIL": os.Getenv("SENDEREMAIL"),
		"SENDERPASS":  os.Getenv("SENDERPASS"),
		"PBS_SECRET":  os.Getenv("PBS_SECRET"),
		"PBS_TOKENID": os.Getenv("PBS_TOKENID"),
		"PVE_SECRET":  os.Getenv("PVE_SECRET"),
		"PVE_TOKENID": os.Getenv("PVE_TOKENID"),
	}
	for key, value := range envVars {
		if value == "" {
			log.Printf("[setup error] %s is not present\n", key)
			return false
		}
	}
	return true
}

func isExternalMounted() bool {
	data, _ := os.ReadFile("/proc/mounts")

	// check if the mount point exists in the data
	return strings.Contains(string(data), "/mnt/external")
}

func IsEverythingConfigured(configPathFlag string) bool {
	log.Printf("[setup info] validating config flag\n")
	if configPathFlag == "" {
		log.Printf("[setup error] please provide the path for the config file\n")
		return false
	}
	log.Printf("[setup info] config flag OK\n")

	log.Printf("[setup info] validating env vars \n")
	if !checkEnvVars() {
		return false
	}
	log.Printf("[setup info] env vars are OK\n")

	log.Printf("[setup info] validating mount point\n")
	if !isExternalMounted() {
		log.Printf("[setup error] mount point is not mounted in the system\n")
		return false
	}
	log.Printf("[setup info] mount point is OK\n")

	return true
}
