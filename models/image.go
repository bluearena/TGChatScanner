package models

import (
	"github.com/jinzhu/gorm"
	"net/url"
	"time"
	"fmt"
)

type Image struct {
	gorm.Model
	Src    string    `json:"src"`
	Tags   []Tag     `gorm:"many2many:images_tags"`
	Date   time.Time `gorm:"not null" json:"date"`
	ChatID int64     `gorm:"not null" json:"-"`
	Chat   Chat      `gorm:"ForeignKey:ChatID;AssociationForeignKey:TGID"`
}

func (img *Image) GetImgByParams(db *gorm.DB, params url.Values) ([]Image, error) {
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
		return nil, db.Error
	}

	if err := db.Error; err != nil {
		return nil, err
	}

	return img_slice, nil
}

func (img *Image) CreateImageWithTags(db *gorm.DB, ts []string) error {
	var tags []Tag
	for _, t := range ts {
		tags = append(tags, Tag{Name: t})
	}

	ch := Chat{
		TGID: img.ChatID,
	}

	tx := db.Begin()
	for _, t := range tags {
		t.Chats = append(t.Chats, ch)
		if err := t.SaveIfUnique(db); err != nil {
			tx.Rollback()
			return fmt.Errorf("unable to save tag: %s", err)
		}
	}
	if err := db.Create(img).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to save image: %s", err)
	}
	if err := db.Model(&img).Association("Tags").Append(tags).Error;
		err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to add img_tag association: %s", err)
	}

	return tx.Commit().Error
}
