package main

import (
	"context"
	"github.com/zhughes3/go-accelerate/internal/pkg/postgres"
	"github.com/zhughes3/go-accelerate/internal/pkg/postgres/pgx"
	"github.com/zhughes3/go-accelerate/internal/pkg/v1/iam"
	"github.com/zhughes3/go-accelerate/internal/pkg/v1/server"
	"github.com/zhughes3/go-accelerate/internal/pkg/v1/user"
	"github.com/zhughes3/go-accelerate/pkg/app"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	"github.com/zhughes3/go-accelerate/pkg/timelines"
	"os"
)

const serviceName = "driver"

func main() {
	config := mustReadEnvConfig()
	logger := mustCreateLogger(config.LoggerConfig)

	db := mustCreateDatabase(logger, config.DBConfig)
	appServer := mustCreateAppServer(logger, config, db)

	appServer.RegisterBeforeShutdownErrorHook(db.Shutdown)

	if err := appServer.Run(context.Background()); err != nil {
		logStaticFatalStartupError("Problem running app server", err)
	}
}

func mustCreateDatabase(logger slog.Logger, config postgres.Config) postgres.DB {
	db, err := pgx.NewDBConnect(context.Background(), logger, &config)
	if err != nil {
		logStaticFatalStartupError("Problem creating database", err)
	}

	return db
}

func mustReadEnvConfig() appConfig {
	config, err := readEnvConfig()
	if err != nil {
		logStaticFatalStartupError("problem reading configuration", err)
	}

	return config
}

func mustCreateLogger(cfg slog.Config) slog.Logger {
	logger, err := slog.
		NewLoggerBuilder().
		WithJSONFormatting().
		WithLevel(cfg.Level).
		WithWriter(os.Stdout).
		Build()
	if err != nil {
		logStaticFatalStartupError("Problem creating logger", err)
	}

	return logger
}

func mustCreateAppServer(logger slog.Logger, cfg appConfig, db postgres.DB) *app.Server {
	logger = logger.WithContextExtractor(user.IDExtractor)

	timelinesService := timelines.NewService(logger, db)

	appServer, err := app.NewServer(logger,
		app.WithPProfEnabled(),
		app.WithAuthMiddleware(iam.NewAuthMiddleware(logger)),
		app.WithRequestHandlerSpecs(server.NewHandlerSpecs(user.MustResolveID, timelinesService)))
	if err != nil {
		logStaticFatalStartupError("Problem creating app server", err)
	}

	return appServer
}

func logStaticFatalStartupError(msg string, err error) {
	logFatalStartupError(slog.Base(), msg, err)
}

func logFatalStartupError(logger slog.Logger, msg string, err error) {
	logger.WithError(err).FatalContext(context.Background(), msg)
}
