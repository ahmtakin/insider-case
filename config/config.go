package config

import (
	"log"
	os "os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB *DBConfig
}
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func LoadConfig() (*Config, error) {

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: No .env file found (this is expected in production)")
		}
	}

	config := &Config{
		DB: &DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		},
	}
	return config, nil
}
