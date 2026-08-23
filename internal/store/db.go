package store

import (
	"context"
	"fmt"
	"time"

	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/models"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config, logger utils.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.DBConfig.Host,
		cfg.DBConfig.User,
		cfg.DBConfig.Password,
		cfg.DBConfig.DBName,
		cfg.DBConfig.Port,
		cfg.DBConfig.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("connection failed : %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	err = db.AutoMigrate(
		&models.User{},
		&models.Webhook{},
		&models.Secret{},
		&models.PipelineRun{},
		&models.PipelineRunStep{},
	)
	if err != nil {
		return nil, fmt.Errorf("migration failed : %w", err)
	}

	// Older installations may contain users created before a default role was
	// enforced. Backfill only missing roles so explicit admin roles are kept.
	rows, err := gorm.G[models.User](db).
		Where(fmt.Sprintf("%s = ? OR %s IS NULL", models.UserColRole, models.UserColRole), "").
		Update(context.Background(), models.UserColRole, models.RoleUser)
	if err != nil {
		return nil, fmt.Errorf("user role backfill failed : %w", err)
	}
	if rows > 0 {
		logger.LogInfo("Backfilled default role for %d user(s)", rows)
	}

	return db, nil
}
