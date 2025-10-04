package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var BuildKey string

type Config struct {
	Privacy PrivacyConfig
	Auth    AuthConfig
	HTTP    HTTPConfig
}

func loadConfig(file string) (Config, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return Config{}, fmt.Errorf("reading config file '%s': %w", file, err)
	}

	var config Config
	if err := json.Unmarshal(raw, &config); err != nil {
		return Config{}, fmt.Errorf("corrupted config file '%s': %w", file, err)
	}
	return config, nil
}
