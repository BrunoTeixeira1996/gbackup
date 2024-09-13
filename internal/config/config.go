package config

import (
	"os"

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

type Target struct {
	Name         string `toml:"name"`
	IP           string `toml:"ip"`
	Keypath      string `toml:"keypath"`
	Instance     string `toml:"instance"`
	MAC          string `toml:"mac"`
	ExternalPath string `toml:"external_path"`
}

type Config struct {
	NAS         NAS         `toml:"nas"`
	Pushgateway Pushgateway `toml:"pushgateway"`
	Targets     []Target    `toml:"targets"`
}

func ReadTomlFile(configPath string) (Config, error) {
	var cfg Config

	input, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	if _, err := toml.Decode(string(input), &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
