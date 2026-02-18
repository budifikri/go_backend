package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/pos-retail/go_backend/internal/config"
)

var DB *gorm.DB

type Database struct {
	*gorm.DB
}

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ensure UUID generator function is available for AutoMigrate defaults.
	// Most models use DEFAULT gen_random_uuid(), which requires the pgcrypto extension.
	if err := ensureUUIDExtension(DB); err != nil {
		return nil, err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxConnections / 2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected successfully!")
	return DB, nil
}

func ensureUUIDExtension(db *gorm.DB) error {
	// Best-effort create pgcrypto; if not permitted, fail early with a clear message.
	_ = db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`).Error

	var probe string
	if err := db.Raw("SELECT gen_random_uuid()::text").Scan(&probe).Error; err == nil && probe != "" {
		return nil
	}

	return fmt.Errorf("missing PostgreSQL function gen_random_uuid(); enable extension with: CREATE EXTENSION IF NOT EXISTS pgcrypto;")
}

func AutoMigrate(models ...interface{}) error {
	return DB.AutoMigrate(models...)
}

func GetDB() *gorm.DB {
	return DB
}

func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func HealthCheck() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
