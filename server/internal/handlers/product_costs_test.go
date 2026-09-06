package handlers_test

import (
	"net/http"
	"testing"

	"warrantykeeper/server/internal/models"
)

func TestCreateProductCost_Success(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/costs", s.token, map[string]any{
		"amount":      450.5,
		"description": "תיקון מדחס",
		"incurred_at": "2026-06-01",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var cost models.ProductCost
	decodeJSON(t, rec, &cost)
	if cost.ProductID != product.ID {
		t.Errorf("ProductID = %v, want %v", cost.ProductID, product.ID)
	}
	if cost.Amount != 450.5 {
		t.Errorf("Amount = %v, want %v", cost.Amount, 450.5)
	}
	if cost.Description != "תיקון מדחס" {
		t.Errorf("Description = %q, want %q", cost.Description, "תיקון מדחס")
	}
	if cost.IncurredAt.Format("2006-01-02") != "2026-06-01" {
		t.Errorf("IncurredAt = %v, want 2026-06-01", cost.IncurredAt)
	}
}

func TestCreateProductCost_DefaultsIncurredAtToToday(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/costs", s.token, map[string]any{
		"amount": 100,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var cost models.ProductCost
	decodeJSON(t, rec, &cost)
	if cost.IncurredAt.IsZero() {
		t.Error("expected IncurredAt to default to now, got zero value")
	}
}

func TestCreateProductCost_MissingAmountReturns400(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/costs", s.token, map[string]any{
		"description": "x",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateProductCost_BadDateFormatReturns400(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/costs", s.token, map[string]any{
		"amount": 100, "incurred_at": "06/01/2026",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateProductCost_ProductFromOtherHouseholdReturns404(t *testing.T) {
	s := newTestSetup(t)
	otherToken, otherHouseholdID := s.createOtherHousehold(t)
	product := models.Product{HouseholdID: otherHouseholdID, Name: "Theirs", Category: "x"}
	if err := s.db.Create(&product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/costs", s.token, map[string]any{"amount": 100})
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var count int64
	s.db.Model(&models.ProductCost{}).Where("product_id = ?", product.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected no cost to be created, found %d", count)
	}

	rec2 := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/costs", otherToken, map[string]any{"amount": 100})
	if rec2.Code != http.StatusCreated {
		t.Errorf("owner's cost: status = %d, want %d", rec2.Code, http.StatusCreated)
	}
}

func TestCreateProductCost_RequiresAuth(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/costs", "", map[string]any{"amount": 100})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListProductCosts_NewestFirst(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/costs", s.token, map[string]any{
		"amount": 100, "description": "ישן", "incurred_at": "2026-01-01",
	})
	doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/costs", s.token, map[string]any{
		"amount": 200, "description": "חדש", "incurred_at": "2026-06-01",
	})

	rec := doJSONAs(t, s.router, http.MethodGet, "/products/"+product.ID.String()+"/costs", s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var costs []models.ProductCost
	decodeJSON(t, rec, &costs)
	if len(costs) != 2 {
		t.Fatalf("got %d costs, want 2", len(costs))
	}
	if costs[0].Description != "חדש" || costs[1].Description != "ישן" {
		t.Errorf("order = [%q, %q], want [\"חדש\", \"ישן\"]", costs[0].Description, costs[1].Description)
	}
}

func TestListProductCosts_EmptyWhenNone(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodGet, "/products/"+product.ID.String()+"/costs", s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var costs []models.ProductCost
	decodeJSON(t, rec, &costs)
	if len(costs) != 0 {
		t.Errorf("got %d costs, want 0", len(costs))
	}
}

func TestListProductCosts_ProductFromOtherHouseholdReturns404(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	otherToken, _ := s.createOtherHousehold(t)
	rec := doJSONAs(t, s.router, http.MethodGet, "/products/"+product.ID.String()+"/costs", otherToken, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
