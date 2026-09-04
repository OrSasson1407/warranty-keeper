package handlers

import (
	"gorm.io/gorm"

	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/ocr"
	"warrantykeeper/server/internal/storage"
)

type Handler struct {
	DB      *gorm.DB
	Cfg     config.Config
	OCR     ocr.Provider
	Storage storage.Store
}

func New(db *gorm.DB, cfg config.Config, ocrProvider ocr.Provider, store storage.Store) *Handler {
	return &Handler{DB: db, Cfg: cfg, OCR: ocrProvider, Storage: store}
}
