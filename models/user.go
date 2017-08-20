package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	ID       uint   `gorm:"primary_key;AUTO_INCREMENT" json:"-"`
	Username string `gorm:"size:64" json:"username"`
	Password string `gorm:"type:varchar(128)" json:"-"`
	Email    string `gorm:"unique" json:"email"`
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
