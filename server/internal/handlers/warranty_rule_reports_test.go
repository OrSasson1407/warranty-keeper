package handlers_test

import (
	"net/http"
	"testing"

	"warrantykeeper/server/internal/models"
)

func TestReportWarrantyRule_Success(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/warranty-report", s.token, map[string]any{
		"note": "המזגן הזה בטח ל-3 שנים לא לשנתיים",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var report models.WarrantyRuleReport
	decodeJSON(t, rec, &report)
	if report.ProductID != product.ID {
		t.Errorf("ProductID = %v, want %v", report.ProductID, product.ID)
	}
	if report.UserID != s.userID {
		t.Errorf("UserID = %v, want %v (the reporting user)", report.UserID, s.userID)
	}
	if report.Note != "המזגן הזה בטח ל-3 שנים לא לשנתיים" {
		t.Errorf("Note = %q, want the submitted note", report.Note)
	}
}

func TestReportWarrantyRule_NoteIsOptional(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/warranty-report", s.token, map[string]any{})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestReportWarrantyRule_ProductFromOtherHouseholdReturns404(t *testing.T) {
	s := newTestSetup(t)
	otherToken, otherHouseholdID := s.createOtherHousehold(t)
	product := models.Product{HouseholdID: otherHouseholdID, Name: "Theirs", Category: "x"}
	if err := s.db.Create(&product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/warranty-report", s.token, map[string]any{})
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var count int64
	s.db.Model(&models.WarrantyRuleReport{}).Where("product_id = ?", product.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected no report to be created, found %d", count)
	}

	rec2 := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/warranty-report", otherToken, map[string]any{})
	if rec2.Code != http.StatusCreated {
		t.Errorf("owner's report: status = %d, want %d", rec2.Code, http.StatusCreated)
	}
}

func TestReportWarrantyRule_RequiresAuth(t *testing.T) {
	s := newTestSetup(t)
	product := seedProduct(t, s)

	rec := doJSONAs(t, s.router, http.MethodPost, "/products/"+product.ID.String()+"/warranty-report", "", map[string]any{})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
