package models

type User struct {
	//gorm.Model
	ID        uint          `gorm:"primary_key;AUTO_INCREMENT"`
	Username  string        `gorm:"size:64"`
	Password  string        `gorm:"type:varchar(128)"`
	Email     string        `gorm:"unique"`
	//Chat []Chat				`gorm:"many2many:users_chats;"`
}
