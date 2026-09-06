package main

import (
	"log"

	"github.com/joho/godotenv"

	"warrantykeeper/server/internal/api"
	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/db"
	"warrantykeeper/server/internal/handlers"
	"warrantykeeper/server/internal/ocr"
	"warrantykeeper/server/internal/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	cfg := config.Load()

	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	var store storage.Store
	if cfg.SupabaseURL != "" && cfg.SupabaseServiceRoleKey != "" {
		store = storage.NewSupabaseStore(cfg.SupabaseURL, cfg.SupabaseStorageBucket, cfg.SupabaseServiceRoleKey)
		log.Printf("using Supabase storage (bucket=%s)", cfg.SupabaseStorageBucket)
	} else {
		localStore, err := storage.NewLocalStore(cfg.UploadsDir, cfg.PublicBaseURL+"/uploads")
		if err != nil {
			log.Fatalf("failed to initialize storage: %v", err)
		}
		store = localStore
		log.Println("using local disk storage (ephemeral -- set SUPABASE_URL/SUPABASE_SERVICE_ROLE_KEY for persistent storage)")
	}

	var ocrProvider ocr.Provider = ocr.NewStubProvider()
	switch {
	case cfg.GeminiAPIKey != "":
		ocrProvider = ocr.NewGeminiProvider(cfg.GeminiAPIKey, cfg.GeminiOCRModel)
		log.Printf("using Gemini OCR provider (model=%s)", cfg.GeminiOCRModel)
	case cfg.AnthropicAPIKey != "":
		ocrProvider = ocr.NewAnthropicProvider(cfg.AnthropicAPIKey, cfg.AnthropicOCRModel)
		log.Printf("using Anthropic OCR provider (model=%s)", cfg.AnthropicOCRModel)
	}

	h := handlers.New(gdb, cfg, ocrProvider, store)
	router := api.NewRouter(h)

	log.Printf("starting server on port %s (env=%s)", cfg.Port, cfg.Env)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
