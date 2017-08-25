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
	Tags   []Tag     `gorm:"many2many:images_tags" json:"tags,omitempty"`
	Date   time.Time `gorm:"not null" json:"date"`
	ChatID int64     `gorm:"not null" json:"chat_id"`
	Chat   Chat      `gorm:"ForeignKey:ChatID;AssociationForeignKey:TGID" json:"-"`
}

func (img *Image) GetImgByParams(db *gorm.DB, params url.Values, user *User) ([]Image, error) {
	imgSlice := []Image{}
	chatIDs := []int64{}
	for _, chat := range user.Chats {
		chatIDs = append(chatIDs, chat.TGID)
	}

	query := db.Model(&Image{}).
		Select("DISTINCT images.*")

	tags, ok := params["tag"]

	if ok {
		query = query.
			Preload("Tags", "name IN (?)", tags).
			Joins("inner join images_tags on images.id = images_tags.image_id inner join tags on images_tags.tag_id = tags.id").
			Where("name in (?)", params["tag"])
	}

	query = query.Where("chat_id in (?) ", chatIDs)

	chat_id, ok := params["chat_id"]
	if ok {
		query = query.Where("chat_id = ?", chat_id[0])
	}
	date_from, ok := params["date_to"]
	if ok {
		query = query.Where("date > ?", date_from[0])
	}
	date_to, ok := params["date_to"]
	if ok {
		query = query.Where("date < ?", date_to[0])
	}

	if query.
		Find(&imgSlice).
		RecordNotFound() {
		return nil, db.Error
	}

	if err := db.Error; err != nil {
		return nil, err
	}

	return imgSlice, nil
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
