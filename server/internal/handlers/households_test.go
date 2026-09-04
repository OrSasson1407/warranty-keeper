package handlers_test

import (
	"net/http"
	"strings"
	"testing"
)

type householdMemberResponse struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type householdResponse struct {
	ID         string                    `json:"id"`
	Name       string                    `json:"name"`
	InviteCode string                    `json:"invite_code"`
	Members    []householdMemberResponse `json:"members"`
}

func TestGetMyHousehold_ReturnsNameInviteCodeAndSoleMember(t *testing.T) {
	s := newTestSetup(t)

	rec := doJSONAs(t, s.router, http.MethodGet, "/households/me", s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	var household householdResponse
	decodeJSON(t, rec, &household)

	if household.ID != s.householdID.String() {
		t.Errorf("ID = %q, want %q", household.ID, s.householdID.String())
	}
	if household.Name == "" {
		t.Error("expected a non-empty household name")
	}
	if household.InviteCode == "" {
		t.Error("expected a non-empty invite code")
	}
	if len(household.Members) != 1 {
		t.Fatalf("got %d members, want 1", len(household.Members))
	}
	if household.Members[0].ID != s.userID.String() {
		t.Errorf("member id = %q, want %q", household.Members[0].ID, s.userID.String())
	}
	if household.Members[0].Email != "owner@example.com" {
		t.Errorf("member email = %q, want %q", household.Members[0].Email, "owner@example.com")
	}
}

func TestGetMyHousehold_ListsAllMembersAndNeverLeaksPasswordHash(t *testing.T) {
	s := newTestSetup(t)
	s.addHouseholdMember(t, "partner@example.com", "Partner")

	rec := doJSONAs(t, s.router, http.MethodGet, "/households/me", s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if strings.Contains(rec.Body.String(), "password_hash") || strings.Contains(rec.Body.String(), "\"x\"") {
		t.Error("response body appears to leak a password hash field")
	}

	var household householdResponse
	decodeJSON(t, rec, &household)
	if len(household.Members) != 2 {
		t.Fatalf("got %d members, want 2", len(household.Members))
	}

	emails := map[string]bool{}
	for _, m := range household.Members {
		emails[m.Email] = true
	}
	if !emails["owner@example.com"] || !emails["partner@example.com"] {
		t.Errorf("members = %+v, want both owner@example.com and partner@example.com", household.Members)
	}
}

func TestGetMyHousehold_ScopedToCallersOwnHousehold(t *testing.T) {
	s := newTestSetup(t)
	otherToken, otherHouseholdID := s.createOtherHousehold(t)

	rec := doJSONAs(t, s.router, http.MethodGet, "/households/me", otherToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var household householdResponse
	decodeJSON(t, rec, &household)

	if household.ID != otherHouseholdID.String() {
		t.Errorf("ID = %q, want the caller's own household %q", household.ID, otherHouseholdID.String())
	}
	if household.ID == s.householdID.String() {
		t.Error("got the wrong household — leaked the first household's data to a different caller")
	}
	if len(household.Members) != 1 {
		t.Errorf("got %d members, want 1 (must not include the first household's members)", len(household.Members))
	}
}

func TestGetMyHousehold_RequiresAuth(t *testing.T) {
	s := newTestSetup(t)
	rec := doJSONAs(t, s.router, http.MethodGet, "/households/me", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
