package models

type Session struct {
	//gorm.Model
	ID	uint 	`gorm:"primary_key;AUTO_INCREMENT"`
	UserID   	uint
	SessionID string
}
