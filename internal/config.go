package internal

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Target struct {
	Name         string `toml:"name"`
	Host         string `toml:"host"`
	Keypath      string `toml:"keypath"`
	Instance     string `toml:"instance"`
	MAC          string `toml:"mac"`
	ExternalPath string `toml:"external_path"`
}

type Pushgateway struct {
	Host string `toml:"host"`
}

type Config struct {
	Pushgateway Pushgateway `toml:"pushgateway"`
	Targets     []Target    `toml:"targets"`
}

func ReadTomlFile() (Config, error) {
	var cfg Config

	//input, err := os.ReadFile("/home/brun0/Desktop/personal/gbackup/config.toml")
	input, err := os.ReadFile("/root/gbackup/config.toml")
	if err != nil {
		return Config{}, err
	}

	if _, err := toml.Decode(string(input), &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
