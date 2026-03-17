package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/akhil-datla/maildruid/internal/config"
	"github.com/akhil-datla/maildruid/internal/domain/user"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps a GORM database connection.
type DB struct {
	db     *gorm.DB
	logger *slog.Logger
}

// New opens a PostgreSQL connection using the provided config.
func New(cfg config.DatabaseConfig, log *slog.Logger) (*DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	log.Info("connected to database", "host", cfg.Host, "name", cfg.Name)
	return &DB{db: db, logger: log}, nil
}

// Migrate runs auto-migrations for all domain models.
func (d *DB) Migrate() error {
	return d.db.AutoMigrate(&user.User{})
}

// Ping checks database connectivity.
func (d *DB) Ping(ctx context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Close closes the database connection.
func (d *DB) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GORM returns the underlying GORM DB for use by repositories.
func (d *DB) GORM() *gorm.DB {
	return d.db
}
