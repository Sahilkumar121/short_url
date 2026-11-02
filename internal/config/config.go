package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DB        string
	SecretKey string
}

func MustLoad() (Config, error) {

	if err := godotenv.Load(); err != nil {
		return Config{}, fmt.Errorf("failed load .env %w", err)
	}

	cfg := Config{
		Port:      os.Getenv("PORT"),
		DB:        os.Getenv("DATABASE"),
		SecretKey: os.Getenv("SECRETKEY"),
	}

	if cfg.Port == "" {
		return Config{}, fmt.Errorf("Port is required")
	}

	if cfg.DB == "" {
		return Config{}, fmt.Errorf("database is required")
	}

	if cfg.SecretKey == "" {
		return Config{}, fmt.Errorf("secrete key is required")
	}

	return cfg, nil
}
