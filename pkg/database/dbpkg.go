package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"lootor_posts/internal/config"
)

var (
	DB         *gorm.DB
	maxRetries = 5
	retryDelay = 2 * time.Second
)

func InitDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable search_path=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port, cfg.Schema,
	)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			DB = db
			log.Printf("Successfully connected to database on attempt %d", attempt)
			return db, nil
		}

		log.Printf("Attempt %d/%d: failed to connect to database: %v", attempt, maxRetries, err)

		if attempt < maxRetries {
			delay := time.Duration(attempt) * retryDelay
			log.Printf("Waiting %v before next attempt...", delay)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
}

func GetDB() *gorm.DB {
	return DB
}

func IsAlive() bool {
	if DB == nil {
		return false
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return false
	}

	return sqlDB.Ping() == nil
}

func Reconnect(cfg *config.DatabaseConfig) error {
	if DB != nil {
		if sqlDB, err := DB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	db, err := InitDB(cfg)
	if err != nil {
		return err
	}

	DB = db
	return nil
}
