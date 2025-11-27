package database

import (
	"fmt"
	"go-poker-arena/internal/logger"
	"go-poker-arena/internal/models"
	"gorm.io/gorm"
)

// MigrationVersion tracks applied migrations
type MigrationVersion struct {
	ID        uint   `gorm:"primarykey"`
	Version   string `gorm:"uniqueIndex;not null"`
	AppliedAt int64  `gorm:"autoCreateTime"`
}

// RunMigrations runs database migrations safely
func RunMigrations(db *gorm.DB, autoMigrate bool) error {
	// Create migrations table if it doesn't exist
	if err := db.AutoMigrate(&MigrationVersion{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	if autoMigrate {
		logger.Warn().Msg("Auto-migrate enabled - use with caution in production")
		return AutoMigrate(db)
	}

	// In production, check for pending migrations
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	logger.Info().Str("version", currentVersion).Msg("Current database version")
	return nil
}

// AutoMigrate runs automatic migrations (development only)
func AutoMigrate(db *gorm.DB) error {
	logger.Info().Msg("Running auto-migrations...")

	models := []interface{}{
		&models.User{},
		&models.Room{},
		&models.Game{},
		&models.GameHistory{},
		&models.PlayerAction{},
		&models.BanRecord{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	// Record migration
	version := "auto_v1"
	var existing MigrationVersion
	if err := db.Where("version = ?", version).First(&existing).Error; err == gorm.ErrRecordNotFound {
		db.Create(&MigrationVersion{Version: version})
	}

	logger.Info().Msg("Auto-migrations completed")
	return nil
}

// getCurrentVersion gets the latest applied migration version
func getCurrentVersion(db *gorm.DB) (string, error) {
	var version MigrationVersion
	err := db.Order("applied_at DESC").First(&version).Error
	if err == gorm.ErrRecordNotFound {
		return "none", nil
	}
	if err != nil {
		return "", err
	}
	return version.Version, nil
}

// Migration represents a database migration
type Migration struct {
	Version string
	Up      func(*gorm.DB) error
	Down    func(*gorm.DB) error
}

// GetMigrations returns all available migrations
func GetMigrations() []Migration {
	return []Migration{
		{
			Version: "001_initial_schema",
			Up: func(db *gorm.DB) error {
				return AutoMigrate(db)
			},
			Down: func(db *gorm.DB) error {
				// Drop all tables
				return db.Migrator().DropTable(
					&models.User{},
					&models.Room{},
					&models.Game{},
					&models.GameHistory{},
					&models.PlayerAction{},
					&models.BanRecord{},
				)
			},
		},
		// Add more migrations here as needed
	}
}

// ApplyMigration applies a specific migration
func ApplyMigration(db *gorm.DB, migration Migration) error {
	// Check if already applied
	var existing MigrationVersion
	if err := db.Where("version = ?", migration.Version).First(&existing).Error; err == nil {
		logger.Info().Str("version", migration.Version).Msg("Migration already applied")
		return nil
	}

	logger.Info().Str("version", migration.Version).Msg("Applying migration")

	// Run migration in transaction
	return db.Transaction(func(tx *gorm.DB) error {
		if err := migration.Up(tx); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		// Record migration
		return tx.Create(&MigrationVersion{Version: migration.Version}).Error
	})
}

// RollbackMigration rolls back a specific migration
func RollbackMigration(db *gorm.DB, migration Migration) error {
	logger.Warn().Str("version", migration.Version).Msg("Rolling back migration")

	return db.Transaction(func(tx *gorm.DB) error {
		if err := migration.Down(tx); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		// Remove migration record
		return tx.Where("version = ?", migration.Version).Delete(&MigrationVersion{}).Error
	})
}
