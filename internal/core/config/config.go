package config

import (
	"github.com/superhorsy/quest-app-backend/internal/core/drivers/psql"
	"github.com/superhorsy/quest-app-backend/internal/core/listeners/http"
)

// AppConfig represents the configuration of our application.
type AppConfig struct {
	HTTP           http.Config `yaml:"http"`
	PSQL           psql.Config `yaml:"psql"`
	PurgeOnRestart bool        `yaml:"purge_on_restart"`
}
