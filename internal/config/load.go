package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/piqnyx/fish-audio-cli/internal/boundedio"
	"github.com/piqnyx/fish-audio-cli/internal/projectpath"
	"github.com/piqnyx/fish-audio-cli/internal/strictjson"
)

const maxConfigFileBytes int64 = 1 << 20

// Load reads a JSON configuration file, applies its values over defaults,
// and resolves the configured Fish API key path to an absolute path.
func Load(paths projectpath.Resolver) (Config, error) {
	path := paths.ConfigPath()
	if path == "" {
		return Config{}, fmt.Errorf(
			"path resolver is not initialized",
		)
	}

	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf(
			"open config %q: %w",
			path,
			err,
		)
	}

	data, readErr := boundedio.ReadAll(
		file,
		maxConfigFileBytes,
	)
	closeErr := file.Close()

	if readErr != nil {
		readErr = fmt.Errorf(
			"read config %q: %w",
			path,
			readErr,
		)
	}

	if closeErr != nil {
		closeErr = fmt.Errorf(
			"close config %q: %w",
			path,
			closeErr,
		)
	}

	if err := errors.Join(
		readErr,
		closeErr,
	); err != nil {
		return Config{}, err
	}

	cfg := Default()

	if err := strictjson.Decode(data, &cfg); err != nil {
		return Config{}, fmt.Errorf(
			"decode config %q: %w",
			path,
			err,
		)
	}

	if err := validateConfigNulls(data); err != nil {
		return Config{}, fmt.Errorf(
			"validate config %q: %w",
			path,
			err,
		)
	}

	fishAPIKeyPath, err := paths.Resolve(
		cfg.Secrets.FishAPIKeyFile,
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
