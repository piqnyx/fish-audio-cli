package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Load reads a JSON configuration file and applies its values over defaults.
func Load(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("config path is empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config %q: %w", path, err)
	}
	defer file.Close()

	cfg := Default()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", path, err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Config{}, fmt.Errorf("config %q contains multiple JSON values", path)
		}

		return Config{}, fmt.Errorf("decode trailing config data %q: %w", path, err)
	}

	return cfg, nil
}
