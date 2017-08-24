package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"net/url"
	"time"
)

type Image struct {
	gorm.Model
	Src    string    `json:"src"`
	Tags   []Tag     `gorm:"many2many:images_tags"`
	Date   time.Time `gorm:"not null" json:"date"`
	ChatID int64     `gorm:"not null" json:"chat_id"`
	Chat   Chat      `gorm:"ForeignKey:ChatID;AssociationForeignKey:TGID" json:"-"`
}

func (img *Image) GetImgByParams(db *gorm.DB, params url.Values, user *User) ([]Image, error) {
	img_slice := []Image{}
	chats_ids := []int64{}
	for _, chat := range user.Chats {
		chats_ids = append(chats_ids, chat.TGID)
	}

	q_tmp := db.Model(&Image{}).
		Select("DISTINCT images.*").
		Preload("Tags", "name IN (?)", params["tag"]).
		Where("chat_id in (?) ", chats_ids)

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

	if q_tmp.
		Joins("inner join images_tags on images.id = images_tags.image_id inner join tags on images_tags.tag_id = tags.id").
		Where("name in (?)", params["tag"]).
		Find(&img_slice).
		RecordNotFound() {
		return nil, db.Error
	}

	if err := db.Error; err != nil {
		return nil, err
	}

	return img_slice, nil
}

func (img *Image) CreateImageWithTags(db *gorm.DB, ts []Tag) error {
	ch := Chat{
		TGID: img.ChatID,
		Tags: ts,
	}

	if err := db.Create(img).Error; err != nil {

		return fmt.Errorf("unable to save image: %s", err)
	}
	for _, t := range ts {
		if err := t.CreateIfUnique(db); err != nil {
			return fmt.Errorf("unable to save tag: %s", err)
		}

		if err := db.Model(&t).Association("Images").Append(img).Error; err != nil {
			return fmt.Errorf("unable to save img-tag: %s", err)
		}
		if err := db.Model(&t).
			Association("Chats").
			Append(&ch).Error; err != nil {
			return fmt.Errorf("unable to save img-tag: %s", err)
		}

	}
	return db.Error
}
