package modelManager

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/zwirec/TGChatScanner/models"
)

func ConnectToDB(dbinfo map[string]interface{}) (*gorm.DB, error) {
	db, err := gorm.Open(dbinfo["engine"].(string),
		fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
			dbinfo["host"],
			dbinfo["port"],
			dbinfo["user"],
			dbinfo["dbname"],
			dbinfo["password"]))
	if err != nil {
		return nil, err
	}
	return db, err
}

func InitDB(db *gorm.DB) {
	db.LogMode(true)
	db.AutoMigrate(&models.User{}, &models.Chat{}, &models.Tag{}, models.Image{}, models.Token{})
	db.AutoMigrate(&models.User_Chat{})
	//db.Model(&models.Token{}).AddForeignKey("user_id", "users(tg_id)", "RESTRICT", "RESTRICT")
	//db.Model(&models.Image{}).AddForeignKey("chat_id", "chats(id)", "RESTRICT", "RESTRICT")
	//db.Model(&models.User_Chat{}).AddForeignKey("chat_id", "chats(id)", "RESTRICT", "RESTRICT")
	//db.Model(&models.User_Chat{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	//db.Model(&models.Token{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	//db.Model(&models.Token{}).AddForeignKey("chat_id", "chats(id)", "RESTRICT", "RESTRICT")
}
