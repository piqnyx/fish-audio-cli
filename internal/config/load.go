package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/piqnyx/fish-audio-cli/internal/projectpath"
)

// Load reads a JSON configuration file, applies its values over defaults,
// and resolves the configured Fish API key path to an absolute path.
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

	fishAPIKeyPath, err := projectpath.Resolve(
		cfg.Secrets.FishAPIKeyFile,
		path,
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"resolve Fish API key path: %w",
			err,
		)
	}

	cfg.Secrets.FishAPIKeyFile = fishAPIKeyPath

	return cfg, nil
}
