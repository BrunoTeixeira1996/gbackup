package config

import (
	"fmt"
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
		return Config{}, fmt.Errorf("[config error] could not read toml file: %s\n", err)
	}

	if _, err := toml.Decode(string(input), &cfg); err != nil {
		return Config{}, fmt.Errorf("[confir error] could not toml decode toml file: %s\n", err)
	}

	return cfg, nil
}
