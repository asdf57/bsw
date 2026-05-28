package db

import (
	"errors"
	"fmt"
	"log"
	"os"

	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/jackc/pgx/v5/pgconn"
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
		&dbmodels.UserDBEntry{},
		&dbmodels.ExchangeDBEntry{},
		&dbmodels.TagDBEntry{},
		&dbmodels.PaymentDBEntry{},
		&dbmodels.DebtDBEntry{},
		&dbmodels.SettlementDBEntry{},
	); err != nil {
		log.Fatalf("auto-migrate failed: %v", err)
	}

	if err := ensurePaymentTagDefaults(db); err != nil {
		log.Fatalf("payment tag initialization failed: %v", err)
	}

	return &BswDB{DB: db}
}

func ensurePaymentTagDefaults(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_name
		ON tags (name)
	`).Error; err != nil {
		return fmt.Errorf("ensure tag name index: %w", err)
	}

	if err := db.Exec(`
		INSERT INTO tags (name, created_at, updated_at)
		VALUES ('general', NOW(), NOW())
		ON CONFLICT (name) DO NOTHING
	`).Error; err != nil {
		return fmt.Errorf("ensure general tag: %w", err)
	}

	if err := db.Exec(`
		INSERT INTO payment_tags (payment_db_entry_id, tag_db_entry_id)
		SELECT payments.id, tags.id
		FROM payments
		CROSS JOIN tags
		WHERE tags.name = 'general'
			AND payments.deleted_at IS NULL
			AND NOT EXISTS (
				SELECT 1
				FROM payment_tags
				WHERE payment_tags.payment_db_entry_id = payments.id
			)
		ON CONFLICT DO NOTHING
	`).Error; err != nil {
		return fmt.Errorf("ensure existing payment tags: %w", err)
	}

	return nil
}

func (b *BswDB) GetUserFromName(name string) (*dbmodels.UserDBEntry, error) {
	var user dbmodels.UserDBEntry

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

	if err := b.DB.Model(&dbmodels.UserDBEntry{}).Where("name = ?", name).Select("id").Take(&uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("failed to find user: %s", err)
			return 0, err
		}

		fmt.Printf("failed to fetch user: %s", err)
		return 0, err
	}

	return uid, nil
}

func (b *BswDB) GetUsersFromNames(names []string) ([]dbmodels.UserDBEntry, error) {
	if len(names) == 0 {
		return []dbmodels.UserDBEntry{}, nil
	}

	var users []dbmodels.UserDBEntry
	if err := b.DB.Where("name IN ?", names).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("get users by names: %w", err)
	}

	return users, nil
}

func (b *BswDB) GetExchangeRatesFromDB() ([]dbmodels.ExchangeDBEntry, error) {
	var exchangeRates []dbmodels.ExchangeDBEntry
	if err := b.DB.Order("date DESC, from_currency, to_currency").Find(&exchangeRates).Error; err != nil {
		return nil, fmt.Errorf("get exchange rates: %w", err)
	}

	return exchangeRates, nil
}

func IsUniqueConstraintError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
