package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	DBUser        string `envconfig:"DB_USER" required:"true"`
	DBPass        string `envconfig:"DB_PASS" required:"true"`
	DBHost        string `envconfig:"DB_HOST" default:"localhost:3306"`
	DBName        string `envconfig:"DB_NAME" required:"true"`
	ServerAddress string `envconfig:"SERVER_ADDRESS" default:":50051"`
}

// Load processes environment variables and returns a Config struct.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
