package models

type Chat struct {
	ID     uint64
	TGID   uint64
	Title  string
	Avatar string
}

type User_Chat struct {
	ID     uint64
	ChatID uint64
	UserID uint64
}

func (User_Chat) TableName() string {
	return "users_chats"
}

//func (u *User_Chat) Validate(db *gorm.DB) bool {
//	var u_ch User_Chat
//	row := db.Where(&u).Find(&u_ch).RowsAffected
//	if row == 1 && time.Now().Before(u_ch.ExpiredTo) {
//		return true
//	} else {
//		return false
//	}
//}
