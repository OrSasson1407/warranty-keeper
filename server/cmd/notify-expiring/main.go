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

	checked, sent, err := notify.RunExpiryCheck(gdb, sender, notify.DefaultWarningDays, time.Now())
	if err != nil {
		log.Fatalf("notify-expiring failed: %v", err)
	}

	log.Printf("notify-expiring complete: %d products checked, %d notifications sent", checked, sent)
}
