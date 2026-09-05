// cmd/demo-data seeds (or removes) a realistic demo/test dataset: one
// household, two users, a handful of products spanning every warranty
// status, a claim, and a receipt. It exists so demo/test data no longer has
// to be created ad hoc via curl + manual psql cleanup against whatever
// database DATABASE_URL happens to point at (see the "prevent demo data
// from being seeded into a database mistaken for production" issue).
//
// Refuses to run unless APP_ENV=development, and only ever touches rows
// belonging to its own demo household (identified by DemoHouseholdName) —
// it never truncates or touches unrelated data.
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"warrantykeeper/server/internal/auth"
	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/db"
	"warrantykeeper/server/internal/models"
	"warrantykeeper/server/internal/warranty"
)

// DemoHouseholdName marks every household this tool creates. Keep it
// distinctive so `-reset` only ever deletes rows it created itself.
const DemoHouseholdName = "[DEMO] הבית של דנה לוי"

const demoPassword = "demopass123"

func main() {
	reset := flag.Bool("reset", false, "remove the demo dataset instead of creating it")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	if cfg.Env != "development" {
		log.Fatalf("refusing to run: APP_ENV=%q, but demo-data only runs when APP_ENV=development "+
			"(this guard exists so demo/test data can't accidentally be seeded into, or wiped from, "+
			"a database that isn't a disposable local dev instance)", cfg.Env)
	}

	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	var household models.Household
	found := gdb.Where("name = ?", DemoHouseholdName).First(&household).Error == nil

	if *reset {
		if !found {
			log.Println("no demo data found, nothing to reset")
			return
		}
		if err := wipeDemoData(gdb, household); err != nil {
			log.Fatalf("failed to reset demo data: %v", err)
		}
		log.Println("demo data removed")
		return
	}

	if found {
		log.Fatalf("demo data already exists (household %q) — run with -reset first to recreate it", DemoHouseholdName)
	}

	if err := seedDemoData(gdb); err != nil {
		log.Fatalf("failed to seed demo data: %v", err)
	}
}

func wipeDemoData(gdb *gorm.DB, household models.Household) error {
	return gdb.Transaction(func(tx *gorm.DB) error {
		var users []models.User
		if err := tx.Where("household_id = ?", household.ID).Find(&users).Error; err != nil {
			return err
		}
		userIDs := make([]uuid.UUID, len(users))
		for i, u := range users {
			userIDs[i] = u.ID
		}

		var products []models.Product
		if err := tx.Where("household_id = ?", household.ID).Find(&products).Error; err != nil {
			return err
		}
		productIDs := make([]uuid.UUID, len(products))
		for i, p := range products {
			productIDs[i] = p.ID
		}

		if len(productIDs) > 0 {
			if err := tx.Where("product_id IN ?", productIDs).Delete(&models.WarrantyClaim{}).Error; err != nil {
				return err
			}
			if err := tx.Where("product_id IN ?", productIDs).Delete(&models.NotificationLog{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("household_id = ?", household.ID).Delete(&models.Product{}).Error; err != nil {
			return err
		}
		if err := tx.Where("household_id = ?", household.ID).Delete(&models.Receipt{}).Error; err != nil {
			return err
		}
		if len(userIDs) > 0 {
			if err := tx.Where("user_id IN ?", userIDs).Delete(&models.DeviceToken{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("household_id = ?", household.ID).Delete(&models.User{}).Error; err != nil {
			return err
		}
		return tx.Delete(&household).Error
	})
}

func seedDemoData(gdb *gorm.DB) error {
	return gdb.Transaction(func(tx *gorm.DB) error {
		household := models.Household{
			Name:       DemoHouseholdName,
			InviteCode: "DEMO" + uuid.New().String()[:4],
		}
		if err := tx.Create(&household).Error; err != nil {
			return err
		}

		passwordHash, err := auth.HashPassword(demoPassword)
		if err != nil {
			return err
		}

		owner := models.User{Email: "demo1@warrantykeeper.app", PasswordHash: passwordHash, FullName: "דנה לוי", HouseholdID: household.ID}
		member := models.User{Email: "demo2@warrantykeeper.app", PasswordHash: passwordHash, FullName: "אורי לוי", HouseholdID: household.ID}
		if err := tx.Create(&owner).Error; err != nil {
			return err
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}
		household.CreatedBy = owner.ID
		if err := tx.Save(&household).Error; err != nil {
			return err
		}

		now := time.Now()
		float := func(v float64) *float64 { return &v }

		products := []models.Product{
			{HouseholdID: household.ID, Name: "מקרר סמסונג", Category: "מקרר", Brand: "Samsung", PurchaseDate: now.AddDate(-1, 0, 0), Price: float(4200), Room: "מטבח", WarrantyExpiresAt: now.AddDate(2, 0, 0)},
			{HouseholdID: household.ID, Name: "מכונת כביסה בוש", Category: "מכונת כביסה", Brand: "Bosch", PurchaseDate: now.AddDate(-1, -6, 0), Price: float(3100), Room: "מרפסת", WarrantyExpiresAt: now.AddDate(2, 0, 0)},
			{HouseholdID: household.ID, Name: "טלוויזיה LG", Category: "טלוויזיה", Brand: "LG", PurchaseDate: now.AddDate(-2, 0, 0), Price: float(2800), Room: "סלון", WarrantyExpiresAt: now.AddDate(0, 0, 15)},
			{HouseholdID: household.ID, Name: "אוזניות אלחוטיות", Category: "אוזניות", Brand: "Apple", PurchaseDate: now.AddDate(-2, 0, 0), Price: float(750), Room: "", WarrantyExpiresAt: now.AddDate(0, 0, -30)},
		}
		for i := range products {
			if err := tx.Create(&products[i]).Error; err != nil {
				return err
			}
		}

		// Fifth product resolves its warranty via the rules engine instead of a manual override.
		vacuumPurchase := now.AddDate(0, -1, 0)
		res, err := warranty.Resolve(tx, "שואב אבק", "Xiaomi", vacuumPurchase)
		if err != nil {
			return err
		}
		vacuum := models.Product{
			HouseholdID: household.ID, Name: "שואב אבק רובוטי", Category: "שואב אבק", Brand: "Xiaomi",
			PurchaseDate: vacuumPurchase, Price: float(1400), Room: "סלון",
			WarrantyExpiresAt: res.ExpiresAt, WarrantyUncertain: res.Uncertain,
		}
		if err := tx.Create(&vacuum).Error; err != nil {
			return err
		}

		// Sixth product demonstrates the receipt-linked path (no photo file backs this demo receipt).
		receipt := models.Receipt{
			HouseholdID: household.ID, Status: models.ReceiptStatusProcessed,
			ParsedVendor: "", RawOCRText: "[demo data: no real OCR output]",
		}
		if err := tx.Create(&receipt).Error; err != nil {
			return err
		}
		mixer := models.Product{
			HouseholdID: household.ID, Name: "מיקסר חשמלי", Category: "מיקרוגל", PurchaseDate: now,
			Room: "מטבח", WarrantyExpiresAt: now.AddDate(1, 0, 0), ReceiptID: &receipt.ID,
		}
		if err := tx.Create(&mixer).Error; err != nil {
			return err
		}

		claim := models.WarrantyClaim{
			ProductID:        products[3].ID, // expired headphones
			IssueDescription: "סוללה לא נטענת יותר",
			Status:           models.ClaimStatusOpen,
		}
		if err := tx.Create(&claim).Error; err != nil {
			return err
		}

		fmt.Println("demo data seeded:")
		fmt.Printf("  household:   %s (invite code %s)\n", household.Name, household.InviteCode)
		fmt.Printf("  users:       %s / %s  (password: %s)\n", owner.Email, member.Email, demoPassword)
		fmt.Printf("  products:    %d (green/orange/red statuses, one rules-engine-resolved, one receipt-linked)\n", len(products)+2)
		fmt.Printf("  claim:       1 open claim on the expired headphones\n")
		fmt.Println("run with -reset to remove this dataset later")
		return nil
	})
}
