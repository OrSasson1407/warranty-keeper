package main

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"warrantykeeper/server/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(
		&models.Household{}, &models.User{}, &models.Product{}, &models.Receipt{},
		&models.WarrantyClaim{}, &models.DeviceToken{}, &models.NotificationLog{},
		&models.WarrantyRule{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestSeedDemoDataCreatesFullDataset(t *testing.T) {
	db := newTestDB(t)

	if err := seedDemoData(db); err != nil {
		t.Fatalf("seedDemoData returned an error: %v", err)
	}

	var household models.Household
	if err := db.Where("name = ?", DemoHouseholdName).First(&household).Error; err != nil {
		t.Fatalf("expected the demo household to exist: %v", err)
	}
	if household.CreatedBy == uuid.Nil {
		t.Errorf("expected household.CreatedBy to be set to the owner's ID")
	}

	var userCount int64
	db.Model(&models.User{}).Where("household_id = ?", household.ID).Count(&userCount)
	if userCount != 2 {
		t.Errorf("expected 2 users, got %d", userCount)
	}

	var productCount int64
	db.Model(&models.Product{}).Where("household_id = ?", household.ID).Count(&productCount)
	if productCount != 6 {
		t.Errorf("expected 6 products, got %d", productCount)
	}

	var receiptCount int64
	db.Model(&models.Receipt{}).Where("household_id = ?", household.ID).Count(&receiptCount)
	if receiptCount != 1 {
		t.Errorf("expected 1 receipt, got %d", receiptCount)
	}

	var claimCount int64
	db.Model(&models.WarrantyClaim{}).
		Joins("JOIN products ON products.id = warranty_claims.product_id").
		Where("products.household_id = ?", household.ID).
		Count(&claimCount)
	if claimCount != 1 {
		t.Errorf("expected 1 claim, got %d", claimCount)
	}
}

func TestSeedDemoDataTwiceFailsOnDuplicateEmails(t *testing.T) {
	// seedDemoData uses fixed demo emails, so calling it twice directly
	// against the same DB fails on the users.email unique constraint rather
	// than silently creating a second household. main() avoids ever hitting
	// this by checking for an existing demo household first (see the
	// "demo data already exists" guard) -- this test just documents that
	// the underlying data layer also refuses the duplicate.
	db := newTestDB(t)

	if err := seedDemoData(db); err != nil {
		t.Fatalf("first seed failed: %v", err)
	}
	if err := seedDemoData(db); err == nil {
		t.Fatal("expected the second seedDemoData call to fail on duplicate emails, got nil error")
	}

	var count int64
	db.Model(&models.Household{}).Where("name = ?", DemoHouseholdName).Count(&count)
	if count != 1 {
		t.Errorf("expected the failed second attempt to leave exactly 1 household (rolled back), got %d", count)
	}
}

func TestWipeDemoDataRemovesEverything(t *testing.T) {
	db := newTestDB(t)
	if err := seedDemoData(db); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	var household models.Household
	if err := db.Where("name = ?", DemoHouseholdName).First(&household).Error; err != nil {
		t.Fatalf("failed to load seeded household: %v", err)
	}

	if err := wipeDemoData(db, household); err != nil {
		t.Fatalf("wipeDemoData returned an error: %v", err)
	}

	var householdCount, userCount, productCount, receiptCount, claimCount int64
	db.Model(&models.Household{}).Where("id = ?", household.ID).Count(&householdCount)
	db.Model(&models.User{}).Where("household_id = ?", household.ID).Count(&userCount)
	db.Model(&models.Product{}).Where("household_id = ?", household.ID).Count(&productCount)
	db.Model(&models.Receipt{}).Where("household_id = ?", household.ID).Count(&receiptCount)
	db.Model(&models.WarrantyClaim{}).
		Joins("JOIN products ON products.id = warranty_claims.product_id").
		Where("products.household_id = ?", household.ID).
		Count(&claimCount)

	for name, got := range map[string]int64{
		"households": householdCount, "users": userCount,
		"products": productCount, "receipts": receiptCount, "claims": claimCount,
	} {
		if got != 0 {
			t.Errorf("expected 0 %s after wipe, got %d", name, got)
		}
	}
}

func TestWipeDemoDataDoesNotTouchOtherHouseholds(t *testing.T) {
	db := newTestDB(t)
	if err := seedDemoData(db); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	other := models.Household{Name: "משק בית אחר לגמרי", InviteCode: "OTHER1"}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("failed to create unrelated household: %v", err)
	}
	otherUser := models.User{Email: "real@example.com", PasswordHash: "x", FullName: "משתמש אמיתי", HouseholdID: other.ID}
	if err := db.Create(&otherUser).Error; err != nil {
		t.Fatalf("failed to create unrelated user: %v", err)
	}

	var demoHousehold models.Household
	if err := db.Where("name = ?", DemoHouseholdName).First(&demoHousehold).Error; err != nil {
		t.Fatalf("failed to load seeded household: %v", err)
	}
	if err := wipeDemoData(db, demoHousehold); err != nil {
		t.Fatalf("wipeDemoData returned an error: %v", err)
	}

	var count int64
	db.Model(&models.Household{}).Where("id = ?", other.ID).Count(&count)
	if count != 1 {
		t.Errorf("expected the unrelated household to survive the wipe, got count %d", count)
	}
	db.Model(&models.User{}).Where("id = ?", otherUser.ID).Count(&count)
	if count != 1 {
		t.Errorf("expected the unrelated user to survive the wipe, got count %d", count)
	}
}
