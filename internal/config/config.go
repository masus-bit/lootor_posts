package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Schema   string
}

type ServerConfig struct {
	Address string
	Port    string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Address: "0.0.0.0:5101"},
		Database: DatabaseConfig{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Database: os.Getenv("POSTGRES_DB"),
			Schema:   os.Getenv("POSTGRES_SCHEMA"),
		},
	}, nil
}
