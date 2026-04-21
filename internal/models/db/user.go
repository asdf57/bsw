package db

import "gorm.io/gorm"

type UserDBEntry struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null"`
}

func (UserDBEntry) TableName() string { return "users" }
