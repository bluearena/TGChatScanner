package models

import "time"

type Chat struct {
	ID     uint64
	TgID   uint64
	Title  string
	Avatar string
	//User   []User   `gorm:"many2many:users_chats;"`
}

type User_Chat struct {
	ChatID uint64
	UserID uint64
	Time   time.Time
}

func (User_Chat) TableName() string {
	return "users_chats"
}
