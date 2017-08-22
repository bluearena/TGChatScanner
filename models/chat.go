package models

import "github.com/jinzhu/gorm"

type Chat struct {
	gorm.Model
	TGID  uint64 `gorm:"primary_key:true" json:"chat_id"`
	Title string `json:"title"`
	Users []User `json:"-"`
	//Image uint
	Images []Image `gorm:"ForeignKey:ChatID;AssociationForeignKey:TGID" json:"images,omitempty"`
}

type User_Chat struct {
	ID     uint64
	ChatID uint64
	UserID uint64
}

func (User_Chat) TableName() string {
	return "users_chats"
}
