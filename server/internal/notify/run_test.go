package notify_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"warrantykeeper/server/internal/models"
	"warrantykeeper/server/internal/notify"
)

type sentMessage struct {
	Token, Title, Body string
}

// fakeSender records every Send call and can be told to fail for specific
// tokens, to test the "don't log a notification if delivery failed" path.
type fakeSender struct {
	calls   []sentMessage
	failFor map[string]bool
}

func (f *fakeSender) Send(_ context.Context, token, title, body string) error {
	f.calls = append(f.calls, sentMessage{token, title, body})
	if f.failFor[token] {
		return errors.New("send failed")
	}
	return nil
}

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
		&models.Household{}, &models.User{}, &models.Product{},
		&models.DeviceToken{}, &models.NotificationLog{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func seedHouseholdAndUser(t *testing.T, db *gorm.DB, email string) models.User {
	t.Helper()
	household := models.Household{Name: "Test Household", InviteCode: strings.ToUpper(strings.ReplaceAll(email, "@", "-"))}
	if err := db.Create(&household).Error; err != nil {
		t.Fatalf("failed to seed household: %v", err)
	}
	user := models.User{Email: email, PasswordHash: "x", FullName: "Test User", HouseholdID: household.ID}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return user
}

func seedProduct(t *testing.T, db *gorm.DB, householdID uuid.UUID, name string, expiresAt time.Time) models.Product {
	t.Helper()
	product := models.Product{
		HouseholdID:       householdID,
		Name:              name,
		Category:          "x",
		PurchaseDate:      expiresAt.AddDate(-1, 0, 0),
		WarrantyExpiresAt: expiresAt,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}
	return product
}

func seedDeviceToken(t *testing.T, db *gorm.DB, userID uuid.UUID, token string) {
	t.Helper()
	dt := models.DeviceToken{UserID: userID, ExpoPushToken: token}
	if err := db.Create(&dt).Error; err != nil {
		t.Fatalf("failed to seed device token: %v", err)
	}
}

func TestRunExpiryCheck_SendsToRegisteredDeviceAndLogsIt(t *testing.T) {
	db := newTestDB(t)
	now := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)
	user := seedHouseholdAndUser(t, db, "a@example.com")
	seedDeviceToken(t, db, user.ID, "token-a")
	expiresAt := now.AddDate(0, 0, notify.DefaultWarningDays).Add(6 * time.Hour) // mid-day on the target date
	product := seedProduct(t, db, user.HouseholdID, "מזגן טורנדו", expiresAt)

	sender := &fakeSender{}
	checked, sent, err := notify.RunExpiryCheck(db, sender, notify.DefaultWarningDays, now)
	if err != nil {
		t.Fatalf("RunExpiryCheck returned error: %v", err)
	}
	if checked != 1 {
		t.Errorf("checked = %d, want 1", checked)
	}
	if sent != 1 {
		t.Errorf("sent = %d, want 1", sent)
	}
	if len(sender.calls) != 1 {
		t.Fatalf("got %d Send calls, want 1", len(sender.calls))
	}
	call := sender.calls[0]
	if call.Token != "token-a" {
		t.Errorf("Send token = %q, want %q", call.Token, "token-a")
	}
	if !strings.Contains(call.Body, product.Name) || !strings.Contains(call.Body, "30") {
		t.Errorf("Send body = %q, want it to mention the product name and %d days", call.Body, notify.DefaultWarningDays)
	}

	var logCount int64
	db.Model(&models.NotificationLog{}).
		Where("user_id = ? AND product_id = ? AND type = ?", user.ID, product.ID, models.NotificationTypeExpiryWarning).
		Count(&logCount)
	if logCount != 1 {
		t.Errorf("notification_log rows = %d, want 1", logCount)
	}
}

func TestRunExpiryCheck_IgnoresProductsOutsideTheExactDayWindow(t *testing.T) {
	db := newTestDB(t)
	now := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)
	user := seedHouseholdAndUser(t, db, "a@example.com")
	seedDeviceToken(t, db, user.ID, "token-a")

	seedProduct(t, db, user.HouseholdID, "Expires in 29 days", now.AddDate(0, 0, 29))
	seedProduct(t, db, user.HouseholdID, "Expires in 31 days", now.AddDate(0, 0, 31))

	sender := &fakeSender{}
	checked, sent, err := notify.RunExpiryCheck(db, sender, notify.DefaultWarningDays, now)
	if err != nil {
		t.Fatalf("RunExpiryCheck returned error: %v", err)
	}
	if checked != 0 || sent != 0 {
		t.Errorf("checked=%d sent=%d, want 0 and 0 (neither product is exactly %d days out)", checked, sent, notify.DefaultWarningDays)
	}
	if len(sender.calls) != 0 {
		t.Errorf("expected no Send calls, got %d", len(sender.calls))
	}
}

func TestRunExpiryCheck_SkipsUserWithNoDeviceToken(t *testing.T) {
	db := newTestDB(t)
	now := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)
	user := seedHouseholdAndUser(t, db, "notoken@example.com")
	// No device token registered for this user.
	seedProduct(t, db, user.HouseholdID, "X", now.AddDate(0, 0, notify.DefaultWarningDays))

	sender := &fakeSender{}
	checked, sent, err := notify.RunExpiryCheck(db, sender, notify.DefaultWarningDays, now)
	if err != nil {
		t.Fatalf("RunExpiryCheck returned error: %v", err)
	}
	if checked != 1 {
		t.Errorf("checked = %d, want 1 (the product is still in the window)", checked)
	}
	if sent != 0 {
		t.Errorf("sent = %d, want 0 (nobody to deliver to)", sent)
	}
}

func TestRunExpiryCheck_SkipsUserAlreadyNotifiedForThisProduct(t *testing.T) {
	db := newTestDB(t)
	now := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)
	user := seedHouseholdAndUser(t, db, "a@example.com")
	seedDeviceToken(t, db, user.ID, "token-a")
	product := seedProduct(t, db, user.HouseholdID, "X", now.AddDate(0, 0, notify.DefaultWarningDays))

	existing := models.NotificationLog{
		UserID: user.ID, ProductID: product.ID,
		Type: models.NotificationTypeExpiryWarning, SentAt: now.AddDate(0, 0, -1),
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("failed to seed existing notification log: %v", err)
	}

	sender := &fakeSender{}
	checked, sent, err := notify.RunExpiryCheck(db, sender, notify.DefaultWarningDays, now)
	if err != nil {
		t.Fatalf("RunExpiryCheck returned error: %v", err)
	}
	if checked != 1 {
		t.Errorf("checked = %d, want 1", checked)
	}
	if sent != 0 {
		t.Errorf("sent = %d, want 0 (already notified)", sent)
	}
	if len(sender.calls) != 0 {
		t.Errorf("expected no Send calls for an already-notified user, got %d", len(sender.calls))
	}

	var logCount int64
	db.Model(&models.NotificationLog{}).Where("product_id = ?", product.ID).Count(&logCount)
	if logCount != 1 {
		t.Errorf("notification_log rows = %d, want still just 1 (no duplicate)", logCount)
	}
}

func TestRunExpiryCheck_MultipleTokensSendToEachButLogOnce(t *testing.T) {
	db := newTestDB(t)
	now := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)
	user := seedHouseholdAndUser(t, db, "a@example.com")
	seedDeviceToken(t, db, user.ID, "token-1")
	seedDeviceToken(t, db, user.ID, "token-2")
	product := seedProduct(t, db, user.HouseholdID, "X", now.AddDate(0, 0, notify.DefaultWarningDays))

	sender := &fakeSender{}
	_, sent, err := notify.RunExpiryCheck(db, sender, notify.DefaultWarningDays, now)
	if err != nil {
		t.Fatalf("RunExpiryCheck returned error: %v", err)
	}
	if sent != 1 {
		t.Errorf("sent = %d, want 1 (one user, even with two devices)", sent)
	}
	if len(sender.calls) != 2 {
		t.Errorf("got %d Send calls, want 2 (one per device)", len(sender.calls))
	}

	var logCount int64
	db.Model(&models.NotificationLog{}).Where("product_id = ?", product.ID).Count(&logCount)
	if logCount != 1 {
		t.Errorf("notification_log rows = %d, want exactly 1 (not one per device)", logCount)
	}
}

func TestRunExpiryCheck_DeliveryFailureDoesNotLogNotification(t *testing.T) {
	db := newTestDB(t)
	now := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)
	user := seedHouseholdAndUser(t, db, "a@example.com")
	seedDeviceToken(t, db, user.ID, "bad-token")
	product := seedProduct(t, db, user.HouseholdID, "X", now.AddDate(0, 0, notify.DefaultWarningDays))

	sender := &fakeSender{failFor: map[string]bool{"bad-token": true}}
	checked, sent, err := notify.RunExpiryCheck(db, sender, notify.DefaultWarningDays, now)
	if err != nil {
		t.Fatalf("RunExpiryCheck returned error: %v", err)
	}
	if checked != 1 {
		t.Errorf("checked = %d, want 1", checked)
	}
	if sent != 0 {
		t.Errorf("sent = %d, want 0 (delivery failed)", sent)
	}

	var logCount int64
	db.Model(&models.NotificationLog{}).Where("product_id = ?", product.ID).Count(&logCount)
	if logCount != 0 {
		t.Errorf("notification_log rows = %d, want 0 so tomorrow's run retries", logCount)
	}
}

func TestRunExpiryCheck_MultipleHouseholdsBothNotified(t *testing.T) {
	db := newTestDB(t)
	now := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)

	userA := seedHouseholdAndUser(t, db, "a@example.com")
	seedDeviceToken(t, db, userA.ID, "token-a")
	seedProduct(t, db, userA.HouseholdID, "Product A", now.AddDate(0, 0, notify.DefaultWarningDays))

	userB := seedHouseholdAndUser(t, db, "b@example.com")
	seedDeviceToken(t, db, userB.ID, "token-b")
	seedProduct(t, db, userB.HouseholdID, "Product B", now.AddDate(0, 0, notify.DefaultWarningDays))

	sender := &fakeSender{}
	checked, sent, err := notify.RunExpiryCheck(db, sender, notify.DefaultWarningDays, now)
	if err != nil {
		t.Fatalf("RunExpiryCheck returned error: %v", err)
	}
	if checked != 2 || sent != 2 {
		t.Errorf("checked=%d sent=%d, want 2 and 2", checked, sent)
	}
}

func TestRunExpiryCheck_NoProductsReturnsZeroes(t *testing.T) {
	db := newTestDB(t)
	sender := &fakeSender{}
	checked, sent, err := notify.RunExpiryCheck(db, sender, notify.DefaultWarningDays, time.Now())
	if err != nil {
		t.Fatalf("RunExpiryCheck returned error: %v", err)
	}
	if checked != 0 || sent != 0 {
		t.Errorf("checked=%d sent=%d, want 0 and 0 on an empty database", checked, sent)
	}
}
