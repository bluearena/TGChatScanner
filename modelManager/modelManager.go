package modelManager

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jinzhu/gorm"
	"fmt"
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
		return nil,err
	}
	return db, err
}
