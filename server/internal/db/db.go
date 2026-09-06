package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"warrantykeeper/server/internal/models"
)

func Connect(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
}

// AutoMigrate creates/updates tables for all entities. Order matters only
// for foreign-key creation on databases that enforce it eagerly; GORM
// handles that for us as long as referenced tables are migrated first.
func AutoMigrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(
		&models.Household{},
		&models.User{},
		&models.Receipt{},
		&models.Product{},
		&models.WarrantyRule{},
		&models.WarrantyClaim{},
		&models.NotificationLog{},
		&models.DeviceToken{},
		&models.ManufacturerContact{},
		&models.ProductCost{},
		&models.WarrantyRuleReport{},
		&models.GmailConnection{},
	)
}
