package main

import (
	"log"

	"github.com/joho/godotenv"

	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/db"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(gdb); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migration complete")
}
