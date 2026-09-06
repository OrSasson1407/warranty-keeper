package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"warrantykeeper/server/internal/models"
)

func TestListManufacturerContactsReturnsAllContactsSortedByBrand(t *testing.T) {
	s := newTestSetup(t)
	for _, c := range []models.ManufacturerContact{
		{Brand: "Samsung", Phone: "*6444", Website: "https://samsung.com"},
		{Brand: "Apple", Phone: "1-800-020-407", Website: "https://apple.com"},
	} {
		if err := s.db.Create(&c).Error; err != nil {
			t.Fatalf("failed to seed manufacturer contact: %v", err)
		}
	}

	rec := doJSONAs(t, s.router, http.MethodGet, "/manufacturer-contacts", s.token, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var contacts []models.ManufacturerContact
	if err := json.Unmarshal(rec.Body.Bytes(), &contacts); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(contacts) != 2 {
		t.Fatalf("expected 2 contacts, got %d", len(contacts))
	}
	if contacts[0].Brand != "Apple" || contacts[1].Brand != "Samsung" {
		t.Errorf("expected contacts sorted by brand (Apple, Samsung), got (%s, %s)", contacts[0].Brand, contacts[1].Brand)
	}
}

func TestListManufacturerContactsReturnsEmptyArrayWhenNoneExist(t *testing.T) {
	s := newTestSetup(t)

	rec := doJSONAs(t, s.router, http.MethodGet, "/manufacturer-contacts", s.token, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var contacts []models.ManufacturerContact
	if err := json.Unmarshal(rec.Body.Bytes(), &contacts); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(contacts) != 0 {
		t.Errorf("expected 0 contacts, got %d", len(contacts))
	}
}

func TestListManufacturerContactsRequiresAuth(t *testing.T) {
	s := newTestSetup(t)

	rec := doJSONAs(t, s.router, http.MethodGet, "/manufacturer-contacts", "", nil)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without a token, got %d", rec.Code)
	}
}
