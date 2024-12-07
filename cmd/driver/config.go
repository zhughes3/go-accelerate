package main

import (
	"context"
	"github.com/sethvargo/go-envconfig"
	"github.com/zhughes3/go-accelerate/internal/pkg/postgres"
	"github.com/zhughes3/go-accelerate/pkg/slog"
)

type appConfig struct {
	LoggerConfig slog.Config     `env:",prefix=LOGGER_"`
	DBConfig     postgres.Config `env:",prefix=DB_"`
}

func readEnvConfig() (appConfig, error) {
	return readAppConfig(envconfig.OsLookuper())
}

func readAppConfig(looker envconfig.Lookuper) (appConfig, error) {
	var config appConfig
	err := envconfig.ProcessWith(context.Background(), &envconfig.Config{
		Target:   &config,
		Lookuper: envconfig.PrefixLookuper("AC_", looker),
	})
	return config, err
}
