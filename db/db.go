package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func CreateDB() *gorm.DB {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("DSN was not set!")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Could not open connection to DB, not good!")
	}

	return db
}
