package models_test

import (
	"testing"

	"github.com/google/uuid"

	"warrantykeeper/server/internal/models"
)

func TestBaseModel_BeforeCreateAssignsAnIDWhenNil(t *testing.T) {
	b := &models.BaseModel{}
	if b.ID != uuid.Nil {
		t.Fatalf("test setup: expected a zero-value BaseModel to start with a nil ID, got %v", b.ID)
	}

	if err := b.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate returned error: %v", err)
	}
	if b.ID == uuid.Nil {
		t.Error("expected BeforeCreate to assign a non-nil ID")
	}
}

func TestBaseModel_BeforeCreateLeavesAnExistingIDAlone(t *testing.T) {
	preset := uuid.New()
	b := &models.BaseModel{ID: preset}

	if err := b.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate returned error: %v", err)
	}
	if b.ID != preset {
		t.Errorf("ID = %v, want the preset value %v to be left unchanged", b.ID, preset)
	}
}

func TestBaseModel_BeforeCreateProducesUniqueIDs(t *testing.T) {
	a, b := &models.BaseModel{}, &models.BaseModel{}
	if err := a.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate returned error: %v", err)
	}
	if err := b.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate returned error: %v", err)
	}
	if a.ID == b.ID {
		t.Error("expected two separate BaseModels to get distinct generated IDs")
	}
}
