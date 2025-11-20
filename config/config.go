package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

const envPrefix = "PARSER_"

type Config struct {
	OsaUrl          string `env:"OSAURL,required"`
	Token           string `env:"TOKEN,required"`
	CredentialsPath string `env:"CREDENTIALS_PATH,required"`
}

func LoadConfig() (*Config, error) {
	config, err := env.ParseAsWithOptions[Config](env.Options{Prefix: envPrefix})
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &config, nil
}
