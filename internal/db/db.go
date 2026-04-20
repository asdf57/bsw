package db

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/asdf57/bsw/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type BswDB struct {
	DB *gorm.DB
}

func NewBswDB() *BswDB {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("DSN was not set!")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Could not open connection to DB, not good!")
	}

	if err := db.AutoMigrate(
		&models.UserDBEntry{},
		&models.ExchangeDBEntry{},
		&models.PaymentDBEntry{},
		&models.DebtDBEntry{},
	); err != nil {
		log.Fatalf("auto-migrate failed: %v", err)
	}

	return &BswDB{DB: db}
}

func (b *BswDB) GetUserFromName(name string) (*models.UserDBEntry, error) {
	var user models.UserDBEntry

	if err := b.DB.Where("name = ?", name).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user %q not found: %w", name, err)
		}
		return nil, fmt.Errorf("get user %q: %w", name, err)
	}

	return &user, nil
}

func (b *BswDB) GetUserIdFromName(name string) (uint, error) {
	var uid uint

	if err := b.DB.Model(&models.UserDBEntry{}).Where("name = ?", name).Select("id").Take(&uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("failed to find user: %s", err)
			return 0, err
		}

		fmt.Printf("failed to fetch user: %s", err)
		return 0, err
	}

	return uid, nil
}

func (b *BswDB) GetUsersFromNames(names []string) ([]models.UserDBEntry, error) {
	if len(names) == 0 {
		return []models.UserDBEntry{}, nil
	}

	var users []models.UserDBEntry
	if err := b.DB.Where("name IN ?", names).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("get users by names: %w", err)
	}

	return users, nil
}
