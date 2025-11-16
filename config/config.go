package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	OsaUrl string `env:"OSAURL,required"`
	Token  string `env:"TOKEN,required"`
}

func LoadConfig() (*Config, error) {
	config, err := env.ParseAsWithOptions[Config](env.Options{Prefix: "PARSER_"})
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &config, nil
}
