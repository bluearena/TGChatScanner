package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Token struct {
	gorm.Model
	Token     string `gorm:"unique"`
	ExpiredTo time.Time

	User   User `gorm:"ForeignKey:UserID"`
	UserID uint
}

func (t *Token) Store(db *gorm.DB) error {
	if db.Create(t).Error != nil {
		return db.Error
	}
	return db.Error
}

func (t *Token) GetUser(db *gorm.DB) *User {
	if !db.Model(&Token{}).Preload("User").Where("token = ? AND expired_to > ?", t.Token, time.Now()).First(t).RecordNotFound() {
		return &t.User
	}
	return nil
}
