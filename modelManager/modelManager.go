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

func InitDB(db *gorm.DB) error {
	db.AutoMigrate(&models.User{}, &models.Chat{}, &models.Tag{}, models.Image{}, models.Token{})
	db.Model(&models.Token{}).AddForeignKey("user_id", "users(tg_id)", "RESTRICT", "RESTRICT")
	return db.Error
}
