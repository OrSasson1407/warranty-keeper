package notify

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"warrantykeeper/server/internal/models"
)

// DefaultWarningDays is how far ahead of expiry the single MVP notification
// fires (see the mvp-scope doc: one flat warning, no multi-tier schedule).
const DefaultWarningDays = 30

// maxSendAttempts bounds the immediate in-process retry for a single push
// send. A transient failure (dropped connection, momentary rate limit) no
// longer has to wait for tomorrow's run before retrying; a persistent
// failure still falls through to that daily retry via the unwritten
// notification_log row, same as before.
const maxSendAttempts = 3

func sendWithRetry(ctx context.Context, sender Sender, token, title, body string) error {
	var lastErr error
	for attempt := 0; attempt < maxSendAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond) // 200ms, then 400ms
		}
		if lastErr = sender.Send(ctx, token, title, body); lastErr == nil {
			return nil
		}
	}
	return lastErr
}

// RunExpiryCheck is the MVP's single scheduled job: find every product whose
// warranty expires in exactly warningDays from now, and push a warning to
// each household member who has a registered device, skipping anyone
// already notified for that product (tracked via notifications_log). `now`
// is a parameter rather than time.Now() so this is deterministically
// testable. Intended to be invoked once daily by an external scheduler
// (cron / Windows Task Scheduler) — no in-process scheduler or queue, per
// the architecture doc's "no queue needed until scale requires it" note.
func RunExpiryCheck(gdb *gorm.DB, sender Sender, warningDays int, now time.Time) (checked, sent int, err error) {
	targetDate := now.AddDate(0, 0, warningDays)
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	var products []models.Product
	if err := gdb.Where("warranty_expires_at >= ? AND warranty_expires_at < ?", startOfDay, endOfDay).
		Find(&products).Error; err != nil {
		return 0, 0, fmt.Errorf("query expiring products: %w", err)
	}

	ctx := context.Background()

	for _, product := range products {
		var users []models.User
		if err := gdb.Where("household_id = ?", product.HouseholdID).Find(&users).Error; err != nil {
			log.Printf("failed to load household members for product %s: %v", product.ID, err)
			continue
		}

		for _, user := range users {
			var alreadySent int64
			gdb.Model(&models.NotificationLog{}).
				Where("user_id = ? AND product_id = ? AND type = ?", user.ID, product.ID, models.NotificationTypeExpiryWarning).
				Count(&alreadySent)
			if alreadySent > 0 {
				continue
			}

			var tokens []models.DeviceToken
			if err := gdb.Where("user_id = ?", user.ID).Find(&tokens).Error; err != nil {
				log.Printf("failed to load device tokens for user %s: %v", user.ID, err)
				continue
			}
			if len(tokens) == 0 {
				continue
			}

			title := "האחריות עומדת לפוג"
			body := fmt.Sprintf("האחריות על %s פגה בעוד %d יום", product.Name, warningDays)

			delivered := false
			for _, token := range tokens {
				if err := sendWithRetry(ctx, sender, token.ExpoPushToken, title, body); err != nil {
					log.Printf("failed to send push to %s after %d attempts: %v", token.ExpoPushToken, maxSendAttempts, err)
					continue
				}
				delivered = true
			}

			if delivered {
				gdb.Create(&models.NotificationLog{
					UserID:    user.ID,
					ProductID: product.ID,
					Type:      models.NotificationTypeExpiryWarning,
					SentAt:    time.Now(),
				})
				sent++
			}
		}
	}

	return len(products), sent, nil
}
