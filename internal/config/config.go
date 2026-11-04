package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DB        string
	SecretKey string
	SMTP      SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

func MustLoad() (Config, error) {

	if err := godotenv.Load(); err != nil {
		return Config{}, fmt.Errorf("failed load .env %w", err)
	}

	// create smtp struct
	smtpCfg := SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}

	// check is all smtp field are provided
	if strings.TrimSpace(smtpCfg.Host) == "" {
		return Config{}, errors.New("smtp host is required")
	}

	if strings.TrimSpace(smtpCfg.Port) == "" {
		return Config{}, errors.New("smtp port is required")
	}

	if strings.TrimSpace(smtpCfg.Username) == "" {
		return Config{}, errors.New("smtp username is required")
	}

	if strings.TrimSpace(smtpCfg.Password) == "" {
		return Config{}, errors.New("smtp password is required")
	}

	// create Config sturct
	cfg := Config{
		Port:      os.Getenv("PORT"),
		DB:        os.Getenv("DATABASE"),
		SecretKey: os.Getenv("SECRETKEY"),
		SMTP:      smtpCfg,
	}

	// check if all field are provided
	if cfg.Port == "" {
		return Config{}, errors.New("Port is required")
	}

	if cfg.DB == "" {
		return Config{}, errors.New("database is required")
	}

	if cfg.SecretKey == "" {
		return Config{}, errors.New("secrete key is required")
	}

	return cfg, nil
}
