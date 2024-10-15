package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
	"github.com/BurntSushi/toml"
)

type NAS struct {
	Name string `toml:"name"`
	IP   string `toml:"ip"`
	MAC  string `toml:"mac"`
}

type Pushgateway struct {
	Url string `toml:"url"`
}

type External struct {
	Name          string         `toml:"name"`
	ExternalPath  string         `toml:"external_path"`
	RsyncCommands []RsyncCommand `toml:"rsync_commands"`
}

type RsyncCommand struct {
	Name    string `toml:"name"`
	Command string `toml:"command"`
}

type Target struct {
	Name          string         `toml:"name"`
	IP            string         `toml:"ip"`
	Keypath       string         `toml:"keypath,omitempty"`
	Instance      string         `toml:"instance"`
	MAC           string         `toml:"mac"`
	ExternalPath  string         `toml:"external_path"`
	RsyncCommands []RsyncCommand `toml:"rsync_commands"`
}

type Config struct {
	NAS         NAS         `toml:"nas"`
	Pushgateway Pushgateway `toml:"pushgateway"`
	External    External    `toml:"external"`
	Targets     []Target    `toml:"targets"`
}

// Helper function to replace {current_time} in RsyncCommand in external
func replaceCurrentTime(commands *[]RsyncCommand, currentTime string) {
	for i := range *commands {
		(*commands)[i].Command = strings.ReplaceAll((*commands)[i].Command, "{current_time}", currentTime)
	}
}

func ReadTomlFile(configPath string) (Config, error) {
	var cfg Config

	input, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("[config error] could not read toml file: %s\n", err)
	}

	if _, err := toml.Decode(string(input), &cfg); err != nil {
		return Config{}, fmt.Errorf("[confir error] could not toml decode toml file: %s\n", err)
	}

	// Replace {current_time} in external commands
	replaceCurrentTime(&cfg.External.RsyncCommands, utils.CurrentTime())

	return cfg, nil
}
