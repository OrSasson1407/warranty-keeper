package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"

	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/db"
	"warrantykeeper/server/internal/gmailsync"
)

// Thin entrypoint: all the actual logic lives in gmailsync.RunScan so it
// can be unit tested without a real scheduler or process exit. Meant to be
// invoked periodically by an external scheduler (OS cron / Task Scheduler),
// same as cmd/notify-expiring.
func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	counts, err := gmailsync.RunScan(gdb, cfg, time.Now())
	if err != nil {
		log.Fatalf("gmail scan failed: %v", err)
	}
	log.Printf("gmail scan complete: %d connections scanned, %d messages matched, %d receipts created",
		counts.ConnectionsScanned, counts.MessagesMatched, counts.ReceiptsCreated)
}
