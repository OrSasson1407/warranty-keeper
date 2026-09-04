package handlers_test

import (
	"net/http"
	"testing"
	"time"

	"warrantykeeper/server/internal/models"
	"warrantykeeper/server/internal/warranty"
)

// NOTE: ListProducts' ?q= search uses Postgres-only "ILIKE" syntax (see
// products.go), which SQLite rejects outright. That's the right call for
// the real Postgres deployment, but it means the search path can't be
// exercised by this in-memory SQLite test harness — there's deliberately no
// test for it here. Covering it would need either a Postgres testcontainer
// or swapping ILIKE for a portable (LOWER(name) LIKE LOWER(?)) query.

func createProduct(t *testing.T, s *productsTestSetup, token string, body map[string]any) (int, models.Product) {
	t.Helper()
	rec := doJSONAs(t, s.router, http.MethodPost, "/products", token, body)
	var product models.Product
	if rec.Code == http.StatusCreated {
		decodeJSON(t, rec, &product)
	}
	return rec.Code, product
}

func TestCreateProduct_ResolvesWarrantyFromSeededRule(t *testing.T) {
	s := newProductsTestSetup(t)
	s.seedRule(t, "מזגן", "", 24)

	code, product := createProduct(t, s, s.token, map[string]any{
		"name":          "מזגן טורנדו",
		"category":      "מזגן",
		"purchase_date": "2026-01-15",
	})
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", code, http.StatusCreated)
	}
	want := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC).AddDate(0, 24, 0)
	if !product.WarrantyExpiresAt.Equal(want) {
		t.Errorf("WarrantyExpiresAt = %v, want %v", product.WarrantyExpiresAt, want)
	}
	if product.WarrantyUncertain {
		t.Error("WarrantyUncertain = true, want false when a rule matched")
	}
	if product.HouseholdID != s.householdID {
		t.Errorf("HouseholdID = %v, want %v", product.HouseholdID, s.householdID)
	}
}

func TestCreateProduct_UnknownCategoryFallsBackAndFlagsUncertain(t *testing.T) {
	s := newProductsTestSetup(t)
	// No rules seeded at all.

	code, product := createProduct(t, s, s.token, map[string]any{
		"name":          "מוצר מסתורי",
		"category":      "קטגוריה שלא קיימת",
		"purchase_date": "2026-01-01",
	})
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", code, http.StatusCreated)
	}
	if !product.WarrantyUncertain {
		t.Error("WarrantyUncertain = false, want true when no rule matches")
	}
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, warranty.DefaultFallbackMonths, 0)
	if !product.WarrantyExpiresAt.Equal(want) {
		t.Errorf("WarrantyExpiresAt = %v, want %v", product.WarrantyExpiresAt, want)
	}
}

func TestCreateProduct_ManualWarrantyOverrideIgnoresRules(t *testing.T) {
	s := newProductsTestSetup(t)
	s.seedRule(t, "מזגן", "", 24) // would resolve to 2028-01-15 if not overridden

	code, product := createProduct(t, s, s.token, map[string]any{
		"name":                "מזגן עם אחריות מוארכת",
		"category":            "מזגן",
		"purchase_date":       "2026-01-15",
		"warranty_expires_at": "2031-01-15",
	})
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", code, http.StatusCreated)
	}
	want := time.Date(2031, 1, 15, 0, 0, 0, 0, time.UTC)
	if !product.WarrantyExpiresAt.Equal(want) {
		t.Errorf("WarrantyExpiresAt = %v, want the manual override %v", product.WarrantyExpiresAt, want)
	}
	if product.WarrantyUncertain {
		t.Error("WarrantyUncertain = true, want false for a manual override")
	}
}

func TestCreateProduct_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body map[string]any
	}{
		{"missing name", map[string]any{"category": "מזגן", "purchase_date": "2026-01-01"}},
		{"missing category", map[string]any{"name": "X", "purchase_date": "2026-01-01"}},
		{"missing purchase_date", map[string]any{"name": "X", "category": "מזגן"}},
		{"bad purchase_date format", map[string]any{"name": "X", "category": "מזגן", "purchase_date": "15/01/2026"}},
		{"bad warranty override format", map[string]any{
			"name": "X", "category": "מזגן", "purchase_date": "2026-01-01", "warranty_expires_at": "not-a-date",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newProductsTestSetup(t)
			code, _ := createProduct(t, s, s.token, tt.body)
			if code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", code, http.StatusBadRequest)
			}
		})
	}
}

func TestCreateProduct_RequiresAuth(t *testing.T) {
	s := newProductsTestSetup(t)
	code, _ := createProduct(t, s, "", map[string]any{
		"name": "X", "category": "מזגן", "purchase_date": "2026-01-01",
	})
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
}

func TestCreateProduct_InheritsPhotoAndLinksFromReceipt(t *testing.T) {
	s := newProductsTestSetup(t)
	receipt := models.Receipt{HouseholdID: s.householdID, ImageURL: "https://fake-storage.test/r1.jpg", Status: models.ReceiptStatusProcessed}
	if err := s.db.Create(&receipt).Error; err != nil {
		t.Fatalf("failed to seed receipt: %v", err)
	}

	code, product := createProduct(t, s, s.token, map[string]any{
		"name": "מוצר מקבלה", "category": "מזגן", "purchase_date": "2026-01-01",
		"receipt_id": receipt.ID.String(),
	})
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", code, http.StatusCreated)
	}
	if product.PhotoURL != receipt.ImageURL {
		t.Errorf("PhotoURL = %q, want inherited %q", product.PhotoURL, receipt.ImageURL)
	}
	if product.ReceiptID == nil || *product.ReceiptID != receipt.ID {
		t.Errorf("ReceiptID = %v, want %v", product.ReceiptID, receipt.ID)
	}
}

func TestCreateProduct_ReceiptFromOtherHouseholdIsRejected(t *testing.T) {
	s := newProductsTestSetup(t)
	_, otherHouseholdID := s.createOtherHousehold(t)
	foreignReceipt := models.Receipt{HouseholdID: otherHouseholdID, ImageURL: "https://fake-storage.test/foreign.jpg"}
	if err := s.db.Create(&foreignReceipt).Error; err != nil {
		t.Fatalf("failed to seed foreign receipt: %v", err)
	}

	code, _ := createProduct(t, s, s.token, map[string]any{
		"name": "X", "category": "מזגן", "purchase_date": "2026-01-01",
		"receipt_id": foreignReceipt.ID.String(),
	})
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (must not link a receipt from another household)", code, http.StatusBadRequest)
	}
}

func TestListProducts_SortedSoonestToExpireFirst(t *testing.T) {
	s := newProductsTestSetup(t)
	createProduct(t, s, s.token, map[string]any{"name": "C", "category": "x", "purchase_date": "2026-01-01", "warranty_expires_at": "2028-01-01"})
	createProduct(t, s, s.token, map[string]any{"name": "A", "category": "x", "purchase_date": "2026-01-01", "warranty_expires_at": "2026-06-01"})
	createProduct(t, s, s.token, map[string]any{"name": "B", "category": "x", "purchase_date": "2026-01-01", "warranty_expires_at": "2027-01-01"})

	rec := doJSONAs(t, s.router, http.MethodGet, "/products", s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var products []models.Product
	decodeJSON(t, rec, &products)
	if len(products) != 3 {
		t.Fatalf("got %d products, want 3", len(products))
	}
	gotOrder := []string{products[0].Name, products[1].Name, products[2].Name}
	want := []string{"A", "B", "C"}
	for i := range want {
		if gotOrder[i] != want[i] {
			t.Errorf("order = %v, want %v", gotOrder, want)
			break
		}
	}
}

func TestListProducts_ScopedToHousehold(t *testing.T) {
	s := newProductsTestSetup(t)
	createProduct(t, s, s.token, map[string]any{"name": "Mine", "category": "x", "purchase_date": "2026-01-01", "warranty_expires_at": "2027-01-01"})

	otherToken, _ := s.createOtherHousehold(t)
	createProduct(t, s, otherToken, map[string]any{"name": "Theirs", "category": "x", "purchase_date": "2026-01-01", "warranty_expires_at": "2027-01-01"})

	rec := doJSONAs(t, s.router, http.MethodGet, "/products", s.token, nil)
	var products []models.Product
	decodeJSON(t, rec, &products)
	if len(products) != 1 || products[0].Name != "Mine" {
		t.Errorf("got %+v, want exactly one product named \"Mine\"", products)
	}
}

func TestGetProduct_NotFoundForOtherHousehold(t *testing.T) {
	s := newProductsTestSetup(t)
	_, product := createProduct(t, s, s.token, map[string]any{"name": "X", "category": "x", "purchase_date": "2026-01-01", "warranty_expires_at": "2027-01-01"})

	otherToken, _ := s.createOtherHousehold(t)
	rec := doJSONAs(t, s.router, http.MethodGet, "/products/"+product.ID.String(), otherToken, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetProduct_Success(t *testing.T) {
	s := newProductsTestSetup(t)
	_, product := createProduct(t, s, s.token, map[string]any{"name": "X", "category": "x", "purchase_date": "2026-01-01", "warranty_expires_at": "2027-01-01"})

	rec := doJSONAs(t, s.router, http.MethodGet, "/products/"+product.ID.String(), s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got models.Product
	decodeJSON(t, rec, &got)
	if got.ID != product.ID {
		t.Errorf("ID = %v, want %v", got.ID, product.ID)
	}
}

func TestUpdateProduct_PartialUpdateAppliesOnlyGivenFields(t *testing.T) {
	s := newProductsTestSetup(t)
	_, product := createProduct(t, s, s.token, map[string]any{"name": "Original", "category": "מזגן", "purchase_date": "2026-01-01", "warranty_expires_at": "2027-01-01"})

	rec := doJSONAs(t, s.router, http.MethodPut, "/products/"+product.ID.String(), s.token, map[string]any{
		"room": "office",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	var updated models.Product
	decodeJSON(t, rec, &updated)
	if updated.Room != "office" {
		t.Errorf("Room = %q, want %q", updated.Room, "office")
	}
	if updated.Name != "Original" {
		t.Errorf("Name = %q, want unchanged %q", updated.Name, "Original")
	}
	if updated.Category != "מזגן" {
		t.Errorf("Category = %q, want unchanged %q", updated.Category, "מזגן")
	}
}

func TestUpdateProduct_WarrantyOverrideClearsUncertainFlag(t *testing.T) {
	s := newProductsTestSetup(t)
	// No rule seeded, so this starts out uncertain.
	_, product := createProduct(t, s, s.token, map[string]any{"name": "X", "category": "לא ידוע", "purchase_date": "2026-01-01"})
	if !product.WarrantyUncertain {
		t.Fatal("expected the freshly created product to start uncertain")
	}

	rec := doJSONAs(t, s.router, http.MethodPut, "/products/"+product.ID.String(), s.token, map[string]any{
		"warranty_expires_at": "2030-01-01",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var updated models.Product
	decodeJSON(t, rec, &updated)
	if updated.WarrantyUncertain {
		t.Error("WarrantyUncertain = true, want false after a manual override")
	}
	want := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	if !updated.WarrantyExpiresAt.Equal(want) {
		t.Errorf("WarrantyExpiresAt = %v, want %v", updated.WarrantyExpiresAt, want)
	}
}

func TestUpdateProduct_NotFoundForOtherHousehold(t *testing.T) {
	s := newProductsTestSetup(t)
	_, product := createProduct(t, s, s.token, map[string]any{"name": "X", "category": "x", "purchase_date": "2026-01-01", "warranty_expires_at": "2027-01-01"})

	otherToken, _ := s.createOtherHousehold(t)
	rec := doJSONAs(t, s.router, http.MethodPut, "/products/"+product.ID.String(), otherToken, map[string]any{"room": "hacked"})
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var stillOriginal models.Product
	s.db.First(&stillOriginal, "id = ?", product.ID)
	if stillOriginal.Room == "hacked" {
		t.Error("a foreign household was able to modify this product")
	}
}
