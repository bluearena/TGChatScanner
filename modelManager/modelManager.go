package modelManager

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jinzhu/gorm"
	"fmt"
)

var db *gorm.DB

func ConnectToDB(dbinfo map[string]string) (err error) {

	if db != nil {
		return nil
	}

	db, err = gorm.Open(dbinfo["engine"],
		fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
			dbinfo["host"],
			dbinfo["port"],
			dbinfo["user"],
			dbinfo["dbname"],
			dbinfo["password"]))
	if err != nil {
		return err
	}

	return nil
}
