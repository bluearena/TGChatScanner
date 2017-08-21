package models

import (
	"github.com/jinzhu/gorm"
	"net/url"
	"time"
)

type Image struct {
	gorm.Model
	Src  string
	Tags []Tag `sql:"many2many:images_tags;"`
	Date time.Time

	Chat   Chat `gorm:"ForeignKey:ChatID"`
	ChatID uint64
}

func (img *Image) GetImgByParams(db *gorm.DB, params url.Values) []Image {
	img_slice := []Image{}
	q_tmp := db.Model(&Image{}).Preload("Tags").Preload("Chat")
	chat_id, ok := params["chat_id"]
	if ok {
		q_tmp = q_tmp.Where("chat_id = ?", chat_id[0])
	}
	date_from, ok := params["date_to"]
	if ok {
		q_tmp = q_tmp.Where("date > ?", date_from[0])
	}
	date_to, ok := params["date_to"]
	if ok {
		q_tmp = q_tmp.Where("date < ?", date_to[0])
	}
	if q_tmp.Find(&img_slice).RecordNotFound() {
		return nil
	}
	return img_slice
}
