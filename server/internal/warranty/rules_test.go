package warranty_test

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"warrantykeeper/server/internal/models"
	"warrantykeeper/server/internal/warranty"
)

// newTestDB gives each test an isolated in-memory SQLite database migrated
// for WarrantyRule. Capping the pool at one connection keeps the whole test
// on a single ":memory:" instance instead of SQLite silently handing out a
// fresh, empty database per connection.
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

	if err := db.AutoMigrate(&models.WarrantyRule{}); err != nil {
		t.Fatalf("failed to migrate WarrantyRule: %v", err)
	}
	return db
}

func seedRule(t *testing.T, db *gorm.DB, category, brand string, months int, source string) {
	t.Helper()
	rule := models.WarrantyRule{Category: category, Brand: brand, DurationMonths: months, Source: source}
	if err := db.Create(&rule).Error; err != nil {
		t.Fatalf("failed to seed rule %q/%q: %v", category, brand, err)
	}
}

func TestResolve_ExactCategoryAndBrandMatchTakesPriority(t *testing.T) {
	db := newTestDB(t)
	seedRule(t, db, "מזגן", "", 12, "default")       // general rule
	seedRule(t, db, "מזגן", "טורנדו", 60, "official") // brand-specific override

	purchase := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	got, err := warranty.Resolve(db, "מזגן", "טורנדו", purchase)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	want := purchase.AddDate(0, 60, 0)
	if !got.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, want)
	}
	if got.DurationMonths != 60 {
		t.Errorf("DurationMonths = %d, want 60", got.DurationMonths)
	}
	if got.Uncertain {
		t.Error("Uncertain = true, want false for an exact brand match")
	}
	if got.Source != "official" {
		t.Errorf("Source = %q, want %q", got.Source, "official")
	}
}

func TestResolve_UnknownBrandFallsBackToGeneralCategoryRule(t *testing.T) {
	db := newTestDB(t)
	seedRule(t, db, "מקרר", "", 24, "default")

	purchase := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	// "סמסונג" has no brand-specific rule, so this must fall back to the
	// general "מקרר" rule rather than the flat default.
	got, err := warranty.Resolve(db, "מקרר", "סמסונג", purchase)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	want := purchase.AddDate(0, 24, 0)
	if !got.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, want)
	}
	if got.Uncertain {
		t.Error("Uncertain = true, want false when a general category rule exists")
	}
	if got.Source != "default" {
		t.Errorf("Source = %q, want %q", got.Source, "default")
	}
}

func TestResolve_NoBrandGivenUsesGeneralRule(t *testing.T) {
	db := newTestDB(t)
	seedRule(t, db, "אוזניות", "", 12, "default")

	purchase := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	got, err := warranty.Resolve(db, "אוזניות", "", purchase)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if got.Uncertain {
		t.Error("Uncertain = true, want false")
	}
	if got.DurationMonths != 12 {
		t.Errorf("DurationMonths = %d, want 12", got.DurationMonths)
	}
}

func TestResolve_UnknownCategoryFallsBackToFlatDefault(t *testing.T) {
	db := newTestDB(t)
	// No rules seeded at all.

	purchase := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got, err := warranty.Resolve(db, "קטגוריה שלא קיימת", "", purchase)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	want := purchase.AddDate(0, warranty.DefaultFallbackMonths, 0)
	if !got.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, want)
	}
	if got.DurationMonths != warranty.DefaultFallbackMonths {
		t.Errorf("DurationMonths = %d, want %d", got.DurationMonths, warranty.DefaultFallbackMonths)
	}
	if !got.Uncertain {
		t.Error("Uncertain = false, want true when nothing matches")
	}
	if got.Source != "fallback" {
		t.Errorf("Source = %q, want %q", got.Source, "fallback")
	}
}

func TestResolve_OtherBrandsRuleDoesNotLeakAcrossBrands(t *testing.T) {
	db := newTestDB(t)
	// Only an LG-specific rule exists for טלוויזיה, and no general rule.
	seedRule(t, db, "טלוויזיה", "LG", 24, "official")

	purchase := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)
	got, err := warranty.Resolve(db, "טלוויזיה", "Samsung", purchase)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if got.DurationMonths != warranty.DefaultFallbackMonths {
		t.Errorf("DurationMonths = %d, want fallback %d (must not reuse LG's rule for Samsung)",
			got.DurationMonths, warranty.DefaultFallbackMonths)
	}
	if !got.Uncertain {
		t.Error("Uncertain = false, want true")
	}
}

func TestResolve_EmptyBrandSkipsBrandLookup(t *testing.T) {
	db := newTestDB(t)
	// A rule keyed to a non-empty brand should never match when the caller
	// passes brand = "" — that must go straight to the general-rule lookup.
	seedRule(t, db, "מחשב נייד", "Dell", 12, "official")

	purchase := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	got, err := warranty.Resolve(db, "מחשב נייד", "", purchase)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if got.Source == "official" {
		t.Error("matched the Dell-specific rule despite an empty brand being requested")
	}
	if !got.Uncertain {
		t.Error("Uncertain = false, want true (no general rule seeded)")
	}
}

func TestResolve_MonthArithmeticFollowsTimeAddDateOverflowRules(t *testing.T) {
	db := newTestDB(t)
	seedRule(t, db, "מוצר קצה", "", 1, "default")

	// Jan 31 + 1 month: February 2026 only has 28 days, so Go's time.AddDate
	// rolls the overflow into March. This test pins down that behavior so a
	// future change to the date math doesn't silently shift it.
	purchase := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	got, err := warranty.Resolve(db, "מוצר קצה", "", purchase)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	want := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)
	if !got.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, want)
	}
}
