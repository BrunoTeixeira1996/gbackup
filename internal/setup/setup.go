package setup

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
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

func setupToml(configPathFlag string) (config.Config, error) {
	var (
		cfg config.Config
		err error
	)

	if cfg, err = config.ReadTomlFile(configPathFlag); err != nil {
		return config.Config{}, fmt.Errorf("[setup error] could read toml file: %s", err)
	}

	return cfg, nil
}

func IsEverythingConfigured(configPathFlag string) (config.Config, bool) {
	var (
		cfg config.Config
		err error
	)

	log.Printf("[setup info] validating config flag\n")
	if configPathFlag == "" {
		log.Printf("[setup error] please provide the path for the config file\n")
		return cfg, false
	}
	log.Printf("[setup info] config flag OK\n")

	log.Printf("[setup info] validating env vars \n")
	if !checkEnvVars() {
		return cfg, false
	}
	log.Printf("[setup info] env vars are OK\n")

	log.Printf("[setup info] validating mount point\n")
	if !isExternalMounted() {
		log.Printf("[setup error] mount point is not mounted in the system\n")
		return cfg, false
	}
	log.Printf("[setup info] mount point is OK\n")

	log.Printf("[setup info] reading toml file\n")
	if cfg, err = setupToml(configPathFlag); err != nil {
		log.Println(err)
		return cfg, false
	}
	log.Printf("[setup info] toml file is OK\n")

	return cfg, true
}
