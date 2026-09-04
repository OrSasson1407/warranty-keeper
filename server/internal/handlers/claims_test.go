package handlers_test

import (
	"net/http"
	"testing"
	"time"

	"warrantykeeper/server/internal/models"
)

func seedProduct(t *testing.T, s *testSetup) models.Product {
	t.Helper()
	product := models.Product{
		HouseholdID:       s.householdID,
		Name:              "Test Product",
		Category:          "x",
		PurchaseDate:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		WarrantyExpiresAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := s.db.Create(&product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}
	return product
}

func TestCreateClaim_Success(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/claims", s.token, map[string]any{
		"issue_description": "המזגן לא מקרר בכלל",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var claim models.WarrantyClaim
	decodeJSON(t, rec, &claim)
	if claim.ProductID != product.ID {
		t.Errorf("ProductID = %v, want %v", claim.ProductID, product.ID)
	}
	if claim.IssueDescription != "המזגן לא מקרר בכלל" {
		t.Errorf("IssueDescription = %q, want %q", claim.IssueDescription, "המזגן לא מקרר בכלל")
	}
	if claim.Status != models.ClaimStatusOpen {
		t.Errorf("Status = %q, want %q", claim.Status, models.ClaimStatusOpen)
	}
	if claim.ResolvedAt != nil {
		t.Error("ResolvedAt should be nil for a freshly created claim")
	}
}

func TestCreateClaim_MissingDescriptionReturns400(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/claims", s.token, map[string]any{})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateClaim_ProductFromOtherHouseholdReturns404(t *testing.T) {
	s := newTestSetup(t)
	otherToken, otherHouseholdID := s.createOtherHousehold(t)
	product := models.Product{
		HouseholdID:       otherHouseholdID,
		Name:              "Their Product",
		Category:          "x",
		PurchaseDate:      time.Now(),
		WarrantyExpiresAt: time.Now().AddDate(1, 0, 0),
	}
	if err := s.db.Create(&product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}

	// The setup's default token belongs to a *different* household than the
	// product, so this must 404 rather than leak the product's existence.
	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/claims", s.token, map[string]any{
		"issue_description": "x",
	})
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var count int64
	s.db.Model(&models.WarrantyClaim{}).Where("product_id = ?", product.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected no claim to be created, found %d", count)
	}

	// Sanity: the rightful owner can still file a claim on it.
	rec2 := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/claims", otherToken, map[string]any{
		"issue_description": "x",
	})
	if rec2.Code != http.StatusCreated {
		t.Errorf("owner's claim: status = %d, want %d", rec2.Code, http.StatusCreated)
	}
}

func TestCreateClaim_RequiresAuth(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/claims", "", map[string]any{
		"issue_description": "x",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListClaims_NewestFirst(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	older := models.WarrantyClaim{ProductID: product.ID, IssueDescription: "Older issue", Status: models.ClaimStatusOpen}
	older.CreatedAt = time.Now().Add(-1 * time.Hour)
	if err := s.db.Create(&older).Error; err != nil {
		t.Fatalf("failed to seed older claim: %v", err)
	}
	newer := models.WarrantyClaim{ProductID: product.ID, IssueDescription: "Newer issue", Status: models.ClaimStatusOpen}
	newer.CreatedAt = time.Now()
	if err := s.db.Create(&newer).Error; err != nil {
		t.Fatalf("failed to seed newer claim: %v", err)
	}

	rec := doJSONAs(t, s.router, http.MethodGet, "/products/"+product.ID.String()+"/claims", s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var claims []models.WarrantyClaim
	decodeJSON(t, rec, &claims)
	if len(claims) != 2 {
		t.Fatalf("got %d claims, want 2", len(claims))
	}
	if claims[0].IssueDescription != "Newer issue" || claims[1].IssueDescription != "Older issue" {
		t.Errorf("order = [%q, %q], want [\"Newer issue\", \"Older issue\"]", claims[0].IssueDescription, claims[1].IssueDescription)
	}
}

func TestListClaims_EmptyWhenNone(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodGet, "/products/"+product.ID.String()+"/claims", s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var claims []models.WarrantyClaim
	decodeJSON(t, rec, &claims)
	if len(claims) != 0 {
		t.Errorf("got %d claims, want 0", len(claims))
	}
}

func TestListClaims_ProductFromOtherHouseholdReturns404(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	otherToken, _ := s.createOtherHousehold(t)
	rec := doJSONAs(t, s.router, http.MethodGet, "/products/"+product.ID.String()+"/claims", otherToken, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
