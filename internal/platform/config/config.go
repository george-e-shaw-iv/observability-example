package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
)

// Config is the type that contains fields that stores the necessary configuration
// gathered from the environment.
type Config struct {
	DaemonPort int `envconfig:"DAEMON_PORT" default:"9000"`
	LogLevel   int `envconfig:"LOG_LEVEL" default:"0"`

	DBUser string `envconfig:"DB_USER" default:"root"`
	DBPass string `envconfig:"DB_PASS" default:"root"`
	DBName string `envconfig:"DB_NAME" default:"list"`
	DBHost string `envconfig:"DB_USER" default:"db"`
	DBPort int    `envconfig:"DB_USER" default:"5432"`

	ReadTimeout     time.Duration `envconfig:"READ_TIMEOUT" default:"5s"`
	WriteTimeout    time.Duration `envconfig:"WRITE_TIMEOUT" default:"10s"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
}

// FromEnvironment collects the configuration from environment variables.
func FromEnvironment() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, errors.Wrap(err, "parse environment variables")
	}

	level := zapcore.Level(cfg.LogLevel)
	if lower, upper := zapcore.DebugLevel, zapcore.FatalLevel; level < lower || level > upper {
		return nil, fmt.Errorf("invalid log level given, must be in range of [%d, %d]", lower, upper)
	}

	return &cfg, nil
}
