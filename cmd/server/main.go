package main

import (
	"context"
	"github.com/cenkalti/backoff/v4"
	"github.com/superhorsy/quest-app-backend/internal/quests"

	"github.com/superhorsy/quest-app-backend/internal/config"
	"github.com/superhorsy/quest-app-backend/internal/core/app"
	"github.com/superhorsy/quest-app-backend/internal/core/drivers/psql"
	"github.com/superhorsy/quest-app-backend/internal/core/listeners/http"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	"github.com/superhorsy/quest-app-backend/internal/events"
	questStore "github.com/superhorsy/quest-app-backend/internal/quests/store"
	httptransport "github.com/superhorsy/quest-app-backend/internal/transport/http"
	"github.com/superhorsy/quest-app-backend/internal/users"
	userStore "github.com/superhorsy/quest-app-backend/internal/users/store"
	"go.uber.org/zap"
)

func main() {
	app.Start(appStart)
}

func appStart(ctx context.Context, a *app.App) ([]app.Listener, error) {
	// Load configuration from config/config.yaml which contains details such as DB connection params
	cfg, err := config.Load(ctx)
	//
	a.Config = *cfg

	if err != nil {
		return nil, err
	}

	// Connect to the postgres DB
	db, err := initDatabase(ctx, cfg, a)
	if err != nil {
		return nil, err
	}

	// Run our migrations which will update the DB or create it if it doesn't exist
	if err := db.MigratePostgres(ctx, "file://migrations"); err != nil {
		return nil, err
	}
	a.OnShutdown(func() {
		if cfg.PurgeOnRestart == true {
			logging.From(ctx).Info("Clearing DB")
			// Temp for development so database is cleared on shutdown
			if err := db.RevertMigrations(ctx, "file://migrations"); err != nil {
				logging.From(ctx).Error("failed to revert migrations", zap.Error(err))
			}
		}
	})

	// Instantiate and connect all our classes
	us := userStore.New(db.GetDB())
	qs := questStore.New(db.GetDB())
	e := events.New()
	u := users.New(us, e)
	q := quests.New(qs, e)

	httpServer := httptransport.New(u, q, db.GetDB())

	// Create an HTTP server
	h, err := http.New(httpServer, cfg.HTTP)
	if err != nil {
		return nil, err
	}

	// Start listening for HTTP requests
	return []app.Listener{
		h,
	}, nil
}

func initDatabase(ctx context.Context, cfg *config.Config, a *app.App) (*psql.Driver, error) {
	db := psql.New(cfg.PSQL)

	err := backoff.Retry(func() error {
		return db.Connect(ctx)
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}

	a.OnShutdown(func() {
		// Shutdown connection when server terminated
		logging.From(ctx).Info("shutting down db connection")
		if err := db.Close(ctx); err != nil {
			logging.From(ctx).Error("failed to close db connection", zap.Error(err))
		}
	})

	return db, nil
}
