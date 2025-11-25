package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerPort     string `envconfig:"SERVER_PORT" default:"8080"`
	DatabaseURL    string `envconfig:"DATABASE_URL" required:"true"`
	MigrationsPath string `envconfig:"MIGRATIONS_PATH" default:"./migrations"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}
