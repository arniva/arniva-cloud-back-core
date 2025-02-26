package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/ismailozdel/core2/utils"
	"github.com/joho/godotenv"
)

type Config struct {
	AppConfig
	DBConfig
}

type AppConfig struct {
	AppPort     string
	AppName     string
	Environment string
	JWTSecret   string
}

type DBConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     string
	SSLMode  string
}

var Cfg *Config

func Load() (*Config, error) {
	// .env dosyasını yükle
	if err := godotenv.Load(); err != nil {
		// .env dosyası yoksa hata verme, varsayılan değerleri kullan
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	appConfig := AppConfig{
		AppPort:     utils.GetEnv("APP_PORT", "8080"),
		AppName:     utils.GetEnv("APP_NAME", "Mikroservis Template"),
		Environment: utils.GetEnv("ENVIRONMENT", "dev"),
		JWTSecret:   utils.GetEnv("JWT_SECRET", "secret"),
	}

	if appConfig.AppPort == "" || appConfig.AppName == "" || appConfig.Environment == "" {
		return nil, errors.New("APP_PORT, APP_NAME, ENVIRONMENT environment variables are required")
	}

	dbConfig := DBConfig{
		Host:     utils.GetEnv("DB_HOST", "localhost"),
		User:     utils.GetEnv("DB_USER", "postgres"),
		Password: utils.GetEnv("DB_PASSWORD", "masterkey"),
		DBName:   utils.GetEnv("DB_NAME", "mikroservis_template"),
		Port:     utils.GetEnv("DB_PORT", "5432"),
		SSLMode:  utils.GetEnv("DB_SSL_MODE", "disable"),
	}

	if dbConfig.Host == "" || dbConfig.User == "" || dbConfig.Password == "" || dbConfig.DBName == "" || dbConfig.Port == "" || dbConfig.SSLMode == "" {
		return nil, errors.New("DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT, DB_SSL_MODE environment variables are required")
	}
	Cfg = &Config{
		AppConfig: appConfig,
		DBConfig:  dbConfig,
	}

	return Cfg, nil
}

func (c *DBConfig) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.Host,
		c.User,
		c.Password,
		c.DBName,
		c.Port,
		c.SSLMode,
	)
}
