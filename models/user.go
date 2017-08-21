package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	TGID     uint64 `gorm:"unique" json:"-"`
	Username string `gorm:"size:64" json:"username"`
}

func (u *User) Register(db *gorm.DB) (int64, error) {
	if err := db.Create(u).Error; err != nil {
		return db.RowsAffected, err
	} else {
		return db.RowsAffected, nil
	}
}

func (u *User) IsExists(db *gorm.DB) bool {
	ok := db.Where(u).First(u).RowsAffected
	if ok == 1 {
		return true
	}
	return false
}

func (u *User) validateToken(db *gorm.DB) {
	token := Token{}
	db.Model(&Token{}).Where("token = ?", token).Related(&token, "Token")
}
