package models

import "time"

type User struct {
	ID uint				`gorm:"primary_key;AUTO_INCREMENT"`
	Username string		`gorm:"size:64"`
	Password string
	Email string		`gorm:"unique"`
	SessionId string
}


type Session struct {
	User User
	SessionId string
}

type Image struct {
	Id  uint	`gorm:"primary_key;AUTO_INCREMENT"`
	Src	string
	Chat_id uint64
	Date time.Time
}

type Chat struct{
	ID uint64
 	TgID uint64
	Title string
	Avatar string
}