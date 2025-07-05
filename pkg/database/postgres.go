package database

import (
	"fmt"
	"log"
	"time"

	"github.com/zaynkorai/enlabs/internal/domain/transaction"
	"github.com/zaynkorai/enlabs/internal/domain/user"
	"github.com/zaynkorai/enlabs/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.SSLMode, cfg.TimeZone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Log GORM queries
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established.")

	if err = runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *gorm.DB) error {
	err := db.AutoMigrate(&user.User{}, &transaction.Transaction{})
	if err != nil {
		return fmt.Errorf("failed to auto migrate database: %w", err)
	}

	log.Println("Database auto-migration completed.")

	predefinedUserIDs := []uint64{1, 2, 3}
	for _, id := range predefinedUserIDs {
		var existingUser user.User
		result := db.First(&existingUser, id)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				newUser := user.User{ID: id}
				if createErr := db.Create(&newUser).Error; createErr != nil {
					return fmt.Errorf("failed to create predefined user %d: %w", id, createErr)
				}
				log.Printf("Predefined user %d created.", id)
			} else {
				return fmt.Errorf("failed to check for predefined user %d: %w", id, result.Error)
			}
		}
	}
	log.Println("Predefined users seeding completed.")

	return nil
}
