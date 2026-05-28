package db

import "gorm.io/gorm"

type UserDBEntry struct {
	gorm.Model
	Name          string `gorm:"uniqueIndex;not null"`
	DiscordHandle string
}

func (UserDBEntry) TableName() string { return "users" }
