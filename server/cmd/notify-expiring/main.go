package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"

	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/db"
	"warrantykeeper/server/internal/notify"
)

// Thin entrypoint: all the actual logic lives in internal/notify.RunExpiryCheck
// so it can be unit tested without a real scheduler or process exit.
func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sender := notify.NewLogSender() // swap for notify.NewExpoSender() once ready to actually deliver
	now := time.Now()

	checked, sent, err := notify.RunMultiTierExpiryCheck(gdb, sender, notify.DefaultWarningDaysTiers, now)
	if err != nil {
		log.Fatalf("notify-expiring failed: %v", err)
	}
	log.Printf("notify-expiring complete: %d product-tier checks, %d notifications sent", checked, sent)

	usersNotified, err := notify.RunAnnualSummary(gdb, sender, now)
	if err != nil {
		log.Fatalf("annual summary failed: %v", err)
	}
	log.Printf("annual summary complete: %d users notified", usersNotified)
}
