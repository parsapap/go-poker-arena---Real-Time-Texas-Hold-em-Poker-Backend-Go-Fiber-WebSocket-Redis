package database

import (
	"fmt"
	"time"

	"go-poker-arena/internal/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// PoolConfig controls the underlying database/sql connection pool.
type PoolConfig struct {
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
}

// Connect opens a PostgreSQL connection using the provided DSN and applies the
// given connection-pool settings. Pooling is essential under load: without
// bounds the server can exhaust database connections or hold idle ones open.
func Connect(dsn string, pool PoolConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Quiet GORM's own logger; we use structured zerolog elsewhere.
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("access underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(pool.ConnMaxLife)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info().
		Int("max_open_conns", pool.MaxOpenConns).
		Int("max_idle_conns", pool.MaxIdleConns).
		Msg("Database connected successfully")
	return db, nil
}
