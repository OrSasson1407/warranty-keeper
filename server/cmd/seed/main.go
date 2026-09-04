package main

import (
	"log"

	"github.com/joho/godotenv"

	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/db"
	"warrantykeeper/server/internal/models"
)

// defaultWarrantyRules is a starter set of general (brand-agnostic) warranty
// durations for common Israeli-household product categories. It intentionally
// covers only the most common ~30 categories for MVP — see mvp-scope doc
// section 6 on the accepted "partial coverage" risk. Extend via the DB
// directly or a future admin tool; this seed only ever adds missing rows.
var defaultWarrantyRules = []models.WarrantyRule{
	// מוצרי חשמל גדולים
	{Category: "מזגן", DurationMonths: 24},
	{Category: "מקרר", DurationMonths: 24},
	{Category: "מקפיא", DurationMonths: 24},
	{Category: "מכונת כביסה", DurationMonths: 24},
	{Category: "מייבש כביסה", DurationMonths: 24},
	{Category: "מדיח כלים", DurationMonths: 24},
	{Category: "תנור בנוי", DurationMonths: 24},
	{Category: "כיריים", DurationMonths: 24},
	{Category: "דוד שמש", DurationMonths: 60},

	// מוצרי חשמל קטנים
	{Category: "מיקרוגל", DurationMonths: 12},
	{Category: "שואב אבק", DurationMonths: 12},
	{Category: "מכונת קפה", DurationMonths: 12},
	{Category: "טוסטר אובן", DurationMonths: 12},
	{Category: "קומקום חשמלי", DurationMonths: 12},
	{Category: "מאוורר", DurationMonths: 12},
	{Category: "מטהר אוויר", DurationMonths: 12},

	// אלקטרוניקה
	{Category: "טלוויזיה", DurationMonths: 24},
	{Category: "מחשב נייד", DurationMonths: 12},
	{Category: "מחשב נייח", DurationMonths: 12},
	{Category: "טאבלט", DurationMonths: 12},
	{Category: "סמארטפון", DurationMonths: 12},
	{Category: "אוזניות", DurationMonths: 12},
	{Category: "רמקול בלוטות'", DurationMonths: 12},
	{Category: "מדפסת", DurationMonths: 12},
	{Category: "מצלמה", DurationMonths: 12},
	{Category: "שעון חכם", DurationMonths: 12},
	{Category: "קונסולת משחקים", DurationMonths: 12},
	{Category: "נתב אינטרנט", DurationMonths: 12},

	// ריהוט
	{Category: "ספה", DurationMonths: 24},
	{Category: "מזרן", DurationMonths: 24},
	{Category: "כיסא משרדי", DurationMonths: 12},
	{Category: "ארון בגדים", DurationMonths: 12},
}

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	created := 0
	for _, rule := range defaultWarrantyRules {
		rule.Source = "default"

		var count int64
		if err := gdb.Model(&models.WarrantyRule{}).
			Where("category = ? AND brand = ?", rule.Category, rule.Brand).
			Count(&count).Error; err != nil {
			log.Fatalf("failed to check existing rule %q: %v", rule.Category, err)
		}
		if count > 0 {
			continue
		}

		if err := gdb.Create(&rule).Error; err != nil {
			log.Fatalf("failed to seed rule %q: %v", rule.Category, err)
		}
		created++
	}

	log.Printf("seed complete: %d new warranty_rules created (%d total in seed set)", created, len(defaultWarrantyRules))
}
