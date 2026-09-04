package warranty

import (
	"time"

	"gorm.io/gorm"

	"warrantykeeper/server/internal/models"
)

// DefaultFallbackMonths is used when no rule matches category or brand at
// all (step 3 of the architecture doc's fallback order).
const DefaultFallbackMonths = 12

// Resolution is the result of resolving a warranty expiry date for a
// category/brand pair, per the architecture doc's 3-step fallback:
//  1. exact category + brand match
//  2. general rule for the category (brand blank)
//  3. a flat default, flagged as uncertain
type Resolution struct {
	ExpiresAt      time.Time
	DurationMonths int
	Uncertain      bool
	Source         string
}

func Resolve(db *gorm.DB, category, brand string, purchaseDate time.Time) (Resolution, error) {
	if brand != "" {
		var rule models.WarrantyRule
		err := db.Where("category = ? AND brand = ?", category, brand).First(&rule).Error
		if err == nil {
			return Resolution{
				ExpiresAt:      purchaseDate.AddDate(0, rule.DurationMonths, 0),
				DurationMonths: rule.DurationMonths,
				Uncertain:      false,
				Source:         rule.Source,
			}, nil
		} else if err != gorm.ErrRecordNotFound {
			return Resolution{}, err
		}
	}

	var rule models.WarrantyRule
	err := db.Where("category = ? AND (brand = '' OR brand IS NULL)", category).First(&rule).Error
	if err == nil {
		return Resolution{
			ExpiresAt:      purchaseDate.AddDate(0, rule.DurationMonths, 0),
			DurationMonths: rule.DurationMonths,
			Uncertain:      false,
			Source:         rule.Source,
		}, nil
	} else if err != gorm.ErrRecordNotFound {
		return Resolution{}, err
	}

	return Resolution{
		ExpiresAt:      purchaseDate.AddDate(0, DefaultFallbackMonths, 0),
		DurationMonths: DefaultFallbackMonths,
		Uncertain:      true,
		Source:         "fallback",
	}, nil
}
