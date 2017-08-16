package modelManager

import (
	"database/sql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jinzhu/gorm"
	"fmt"
	"log"
)

var db *sql.DB



func ConnectToDB(dbinfo map[string]interface{}) error {
	db, err := gorm.Open("postgres",
							fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
										dbinfo["host"],
										dbinfo["port"],
										dbinfo["user"],
										dbinfo["dbname"],
										dbinfo["password"]))
	if err != nil {
		return err
	}

	log.Print(db)
	return nil
}
