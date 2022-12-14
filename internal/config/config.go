package config

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
	"os"

	"github.com/superhorsy/quest-app-backend/internal/core/config"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	"go.uber.org/zap"
)

const (
	// ErrInvalidEnvironment is returned when the ENVIRONMENT variable is not set.
	ErrInvalidEnvironment = errors.Error("ENVIRONMENT is not set")
	// ErrValidation is returned when the configuration is invalid.
	ErrValidation = errors.Error("invalid configuration")
	// ErrEnvVars is returned when the environment variables are invalid.
	ErrEnvVars = errors.Error("failed parsing env vars")
	// ErrRead is returned when the configuration file cannot be read.
	ErrRead = errors.Error("failed to read file")
	// ErrUnmarshal is returned when the configuration file cannot be unmarshalled.
	ErrUnmarshal = errors.Error("failed to unmarshal file")
)

var (
	baseConfigPath = "config/config.yaml"
	envConfigPath  = "config/config-%s.yaml"
)

// Config represents the configuration of our application.
type Config struct {
	config.AppConfig `yaml:",inline"`
}

// Load loads the configuration from the config/config.yaml file.
func Load(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	// First load from config/config-*.yaml files
	if err := loadFromFiles(ctx, cfg); err != nil {
		return nil, err
	}
	// Then fill struct with environment values
	if err := env.Parse(cfg); err != nil {
		return nil, ErrEnvVars.Wrap(err)
	}

	logging.From(ctx).Info(fmt.Sprintf("Config: %v.\n", cfg))

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, ErrValidation.Wrap(err)
	}

	return cfg, nil
}

func loadFromFiles(ctx context.Context, cfg any) error {
	environ := os.Getenv("ENVIRONMENT")
	if environ == "" {
		return ErrInvalidEnvironment
	}

	if err := loadYaml(ctx, baseConfigPath, cfg); err != nil {
		return err
	}

	p := fmt.Sprintf(envConfigPath, environ)

	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		if err := loadYaml(ctx, p, cfg); err != nil {
			return err
		}
	}

	return nil
}

func loadYaml(ctx context.Context, filename string, cfg any) error {
	logging.From(ctx).Info("Loading configuration", zap.String("path", filename))

	data, err := os.ReadFile(filename)
	if err != nil {
		return ErrRead.Wrap(err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return ErrUnmarshal.Wrap(err)
	}

	return nil
}
