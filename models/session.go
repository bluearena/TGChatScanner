package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Token struct {
	ID        uint64 `gorm:"primary_key;AUTO_INCREMENT"`
	UserID    uint64
	ChatID    uint64
	Token     string
	ExpiredTo time.Time
}

func (s *Token) Store(db *gorm.DB) error {
	if db.Create(s).Error != nil {
		return db.Error
	}
	return db.Error
}
