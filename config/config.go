package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

const envPrefix = "PARSER_"

type Config struct {
	OsaUrl string `env:"OSAURL,required"`
	Token  string `env:"TOKEN,required"`

	MaxRetries int           `env:"RETRIES" envDefault:"3"`
	RetryDelay time.Duration `env:"RETRY_DELAY" envDefault:"2s"`

	Interval time.Duration `env:"INTERVAL" envDefault:"1h"`
}

func LoadConfig() (*Config, error) {
	config, err := env.ParseAsWithOptions[Config](env.Options{Prefix: envPrefix})
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &config, nil
}
