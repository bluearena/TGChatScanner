package models

import (
	"github.com/jinzhu/gorm"
)

type Session struct {
	ID        uint `gorm:"primary_key;AUTO_INCREMENT"`
	UserID    uint
	SessionID string
}

func (s *Session) Store(db *gorm.DB) error {
	if db.Create(s).Error != nil {
		return db.Error
	}
	return db.Error
}
