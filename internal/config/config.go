package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DB        string
	SecretKey string
	SMTP      SMTPConfig
	Redis     RedisConfig
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

type RedisConfig struct {
	Addr     string
	Password string
	RedisDB  int
}

func MustLoad() (Config, error) {

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("failed load .env %w", err)
	}

	// smtp

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	// check is all smtp field are provided
	if strings.TrimSpace(smtpHost) == "" {
		return Config{}, errors.New("smtp host is required")
	}

	if strings.TrimSpace(smtpPort) == "" {
		return Config{}, errors.New("smtp port is required")
	}

	if strings.TrimSpace(smtpUsername) == "" {
		return Config{}, errors.New("smtp username is required")
	}

	if strings.TrimSpace(smtpPassword) == "" {
		return Config{}, errors.New("smtp password is required")
	}

	// create smtp struct
	smtpCfg := SMTPConfig{
		Host:     smtpHost,
		Port:     smtpPort,
		Username: smtpUsername,
		Password: smtpPassword,
	}

	// redis
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPass := os.Getenv("REDIS_PASS")
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))

	// check redis config
	if strings.TrimSpace(redisAddr) == "" {
		return Config{}, errors.New("redis address is required")
	}

	if err != nil || redisDB < 0 {
		return Config{}, errors.New("failed to retrieve redis DB")
	}

	// create redis config
	redisConfig := RedisConfig{
		Addr:     redisAddr,
		Password: redisPass,
		RedisDB:  redisDB,
	}

	// server
	serverPort := os.Getenv("PORT")

	// database
	postgresqlDB := os.Getenv("DATABASE")

	// password secrete key
	secreteKey := os.Getenv("SECRETKEY")

	// check if all field are provided
	if strings.TrimSpace(serverPort) == "" {
		return Config{}, errors.New("Port is required")
	}

	if strings.TrimSpace(postgresqlDB) == "" {
		return Config{}, errors.New("database is required")
	}

	if strings.TrimSpace(secreteKey) == "" {
		return Config{}, errors.New("secrete key is required")
	}

	// create Config struct
	cfg := Config{
		Port:      serverPort,
		DB:        postgresqlDB,
		SecretKey: secreteKey,
		SMTP:      smtpCfg,
		Redis:     redisConfig,
	}

	return cfg, nil
}
