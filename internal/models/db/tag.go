package db

import (
	"gorm.io/gorm"
)

type TagDBEntry struct {
	gorm.Model
	Name string `gorm:"uniqueIndex"`
}

func (TagDBEntry) TableName() string { return "tags" }
