package dbManager

import (
	"github.com/jinzhu/gorm"
	"fmt"
)

type DBManager struct {
	db *gorm.DB
}

func NewDBManager(dbinfo map[string]string) (*DBManager, error) {
	db, err := gorm.Open("postgres",
		fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
			dbinfo["host"],
			dbinfo["port"],
			dbinfo["user"],
			dbinfo["dbname"],
			dbinfo["password"]))
	if err != nil {
		return nil, err
	}
	return &DBManager{db: db}, nil
}
