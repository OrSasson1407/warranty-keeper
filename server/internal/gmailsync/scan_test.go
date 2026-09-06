package gmailsync

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/crypto"
	"warrantykeeper/server/internal/models"
)

// fakeMessageSource lets a test control exactly what "Gmail" returns
// without a real inbox or network call.
type fakeMessageSource struct {
	ids      []string
	messages map[string]candidateMessage
	fetchErr error
}

func (f *fakeMessageSource) ListCandidates(context.Context) ([]string, error) {
	return f.ids, nil
}

func (f *fakeMessageSource) FetchMessage(_ context.Context, id string) (candidateMessage, error) {
	if f.fetchErr != nil {
		return candidateMessage{}, f.fetchErr
	}
	return f.messages[id], nil
}

func newScanTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := gdb.AutoMigrate(&models.Household{}, &models.User{}, &models.Receipt{}, &models.GmailConnection{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return gdb
}

func seedConnectedUser(t *testing.T, gdb *gorm.DB, cfg config.Config) (models.User, models.GmailConnection) {
	t.Helper()
	household := models.Household{Name: "Test Household", InviteCode: "TEST1234"}
	if err := gdb.Create(&household).Error; err != nil {
		t.Fatalf("failed to seed household: %v", err)
	}
	user := models.User{Email: "owner@example.com", PasswordHash: "x", HouseholdID: household.ID}
	if err := gdb.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	encAccess, _ := crypto.Encrypt("access-token", cfg.TokenEncryptionKey)
	encRefresh, _ := crypto.Encrypt("refresh-token", cfg.TokenEncryptionKey)
	conn := models.GmailConnection{
		UserID: user.ID, GmailAddress: "owner@gmail.com",
		EncryptedAccessToken: encAccess, EncryptedRefreshToken: encRefresh,
		TokenExpiry: time.Now().Add(time.Hour),
	}
	if err := gdb.Create(&conn).Error; err != nil {
		t.Fatalf("failed to seed gmail connection: %v", err)
	}
	return user, conn
}

func TestRunScan_CreatesReceiptsForAllowlistedMatches(t *testing.T) {
	cfg := config.Config{TokenEncryptionKey: "test-key"}
	gdb := newScanTestDB(t)
	user, _ := seedConnectedUser(t, gdb, cfg)

	original := newGmailSource
	defer func() { newGmailSource = original }()
	newGmailSource = func(context.Context, oauth2.TokenSource) (messageSource, error) {
		return &fakeMessageSource{
			ids: []string{"msg-amazon", "msg-unrelated"},
			messages: map[string]candidateMessage{
				"msg-amazon": {
					ID: "msg-amazon", From: "Amazon <auto-confirm@amazon.com>",
					Subject: "Your order has shipped", Date: time.Now(), Body: "Total: ₪199.90",
				},
				"msg-unrelated": {
					ID: "msg-unrelated", From: "noreply@some-newsletter.com",
					Subject: "Weekly digest", Date: time.Now(), Body: "",
				},
			},
		}, nil
	}

	counts, err := RunScan(gdb, cfg, time.Now())
	if err != nil {
		t.Fatalf("RunScan returned error: %v", err)
	}
	if counts.ConnectionsScanned != 1 {
		t.Errorf("ConnectionsScanned = %d, want 1", counts.ConnectionsScanned)
	}
	if counts.MessagesMatched != 1 {
		t.Errorf("MessagesMatched = %d, want 1 (only the allowlisted sender)", counts.MessagesMatched)
	}
	if counts.ReceiptsCreated != 1 {
		t.Errorf("ReceiptsCreated = %d, want 1", counts.ReceiptsCreated)
	}

	var receipts []models.Receipt
	if err := gdb.Find(&receipts).Error; err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(receipts) != 1 {
		t.Fatalf("expected exactly one receipt row, got %d", len(receipts))
	}
	r := receipts[0]
	if r.HouseholdID != user.HouseholdID {
		t.Errorf("HouseholdID = %v, want %v", r.HouseholdID, user.HouseholdID)
	}
	if r.Source != models.ReceiptSourceGmail {
		t.Errorf("Source = %q, want %q", r.Source, models.ReceiptSourceGmail)
	}
	if r.GmailMessageID != "msg-amazon" {
		t.Errorf("GmailMessageID = %q, want %q", r.GmailMessageID, "msg-amazon")
	}
	if r.ParsedVendor != "Amazon" {
		t.Errorf("ParsedVendor = %q, want %q", r.ParsedVendor, "Amazon")
	}
	if r.ParsedAmount == nil || *r.ParsedAmount != 199.90 {
		t.Errorf("ParsedAmount = %v, want %v", r.ParsedAmount, 199.90)
	}
}

func TestRunScan_SkipsAlreadyProcessedMessages(t *testing.T) {
	cfg := config.Config{TokenEncryptionKey: "test-key"}
	gdb := newScanTestDB(t)
	user, _ := seedConnectedUser(t, gdb, cfg)

	if err := gdb.Create(&models.Receipt{
		HouseholdID: user.HouseholdID, Status: models.ReceiptStatusPending,
		Source: models.ReceiptSourceGmail, GmailMessageID: "msg-amazon",
	}).Error; err != nil {
		t.Fatalf("failed to seed existing receipt: %v", err)
	}

	original := newGmailSource
	defer func() { newGmailSource = original }()
	fetchCalled := false
	newGmailSource = func(context.Context, oauth2.TokenSource) (messageSource, error) {
		return &fakeMessageSourceWithFetchTracking{&fakeMessageSource{
			ids: []string{"msg-amazon"},
			messages: map[string]candidateMessage{
				"msg-amazon": {ID: "msg-amazon", From: "auto-confirm@amazon.com", Subject: "x", Date: time.Now()},
			},
		}, &fetchCalled}, nil
	}

	counts, err := RunScan(gdb, cfg, time.Now())
	if err != nil {
		t.Fatalf("RunScan returned error: %v", err)
	}
	if counts.ReceiptsCreated != 0 {
		t.Errorf("ReceiptsCreated = %d, want 0 for an already-processed message", counts.ReceiptsCreated)
	}
	if fetchCalled {
		t.Error("expected FetchMessage not to be called for a message already recorded as processed")
	}
}

// fakeMessageSourceWithFetchTracking wraps fakeMessageSource to record
// whether FetchMessage was ever called, so a test can assert the dedup
// check short-circuits before fetching the full message body.
type fakeMessageSourceWithFetchTracking struct {
	*fakeMessageSource
	fetchCalled *bool
}

func (f *fakeMessageSourceWithFetchTracking) FetchMessage(ctx context.Context, id string) (candidateMessage, error) {
	*f.fetchCalled = true
	return f.fakeMessageSource.FetchMessage(ctx, id)
}

func TestRunScan_UpdatesLastScanAt(t *testing.T) {
	cfg := config.Config{TokenEncryptionKey: "test-key"}
	gdb := newScanTestDB(t)
	_, conn := seedConnectedUser(t, gdb, cfg)

	original := newGmailSource
	defer func() { newGmailSource = original }()
	newGmailSource = func(context.Context, oauth2.TokenSource) (messageSource, error) {
		return &fakeMessageSource{ids: []string{}}, nil
	}

	now := time.Now()
	if _, err := RunScan(gdb, cfg, now); err != nil {
		t.Fatalf("RunScan returned error: %v", err)
	}

	var updated models.GmailConnection
	if err := gdb.First(&updated, "id = ?", conn.ID).Error; err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if updated.LastScanAt == nil {
		t.Fatal("expected LastScanAt to be set after a scan")
	}
	if !updated.LastScanAt.Equal(now) {
		t.Errorf("LastScanAt = %v, want %v", updated.LastScanAt, now)
	}
}
