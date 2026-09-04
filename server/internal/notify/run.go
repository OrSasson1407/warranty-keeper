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
				if err := sender.Send(ctx, token.ExpoPushToken, title, body); err != nil {
					log.Printf("failed to send push to %s: %v", token.ExpoPushToken, err)
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
