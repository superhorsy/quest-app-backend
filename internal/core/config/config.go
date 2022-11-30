package config

import (
	"github.com/superhorsy/quest-app-backend/internal/core/drivers/psql"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/listeners/http"
)

var config *AppConfig = nil

// AppConfig represents the configuration of our application.
type AppConfig struct {
	HTTP           http.Config `yaml:"http"`
	PSQL           psql.Config `yaml:"psql"`
	PurgeOnRestart bool        `yaml:"purge_on_restart"`
	JwtPrivateKey  string      `env:"JWT_PRIVATE_KEY" validate:"required"`
}

func (*AppConfig) Set(appConfig AppConfig) {
	config = &appConfig
}
func (*AppConfig) Get() (*AppConfig, error) {
	if config != nil {
		return nil, errors.New("No config found")
	}
	return config, nil
}

type ContextKey string
