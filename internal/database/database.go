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
	if err := DB.AutoMigrate(models...); err != nil {
		return err
	}
	if err := ensureAppointmentConstraints(DB); err != nil {
		return err
	}
	// Best-effort backfill for master is_active fields.
	_ = BackfillMasterIsActive(DB)
	return nil
}

func ensureAppointmentConstraints(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	stmts := []string{
		`ALTER TABLE appointments DROP CONSTRAINT IF EXISTS fk_appointments_treatment;`,
		`ALTER TABLE appointments ADD CONSTRAINT fk_appointments_treatment FOREIGN KEY (treatment_id) REFERENCES treatments(id);`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

// BackfillMasterIsActive sets is_active from legacy status columns.
// Master tables only; transaction tables keep multi-status.
func BackfillMasterIsActive(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	stmts := []string{
		`UPDATE users SET is_active = (LOWER(status) = 'active') WHERE status IS NOT NULL;`,
		`UPDATE customers SET is_active = (LOWER(status) = 'active') WHERE status IS NOT NULL;`,
		`UPDATE suppliers SET is_active = (LOWER(status) = 'active') WHERE status IS NOT NULL;`,
		`UPDATE warehouses SET is_active = (LOWER(status) = 'active') WHERE status IS NOT NULL;`,
		`UPDATE products SET is_active = (LOWER(status) = 'active') WHERE status IS NOT NULL;`,
		`UPDATE companies SET is_active = (LOWER(status) = 'active') WHERE status IS NOT NULL;`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			// Don't fail startup for backfill issues; migrations may run on partial schemas in tests.
			continue
		}
	}
	return nil
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
