package models

type Session struct {
	//gorm.Model
	User User
	UserID uint
	SessionID string
}
