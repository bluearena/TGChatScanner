package models

import (
	"github.com/jinzhu/gorm"
	"net/url"
	"time"
)

type Image struct {
	gorm.Model
	Src    string    `json:"src"`
	Tags   []Tag     `gorm:"many2many:images_tags"`
	Date   time.Time `gorm:"not null" json:"date"`
	ChatID int64    `gorm:"not null" json:"-"`
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

//TODO: call exorcist on this pile of unholy shit
func (img *Image) CreateImageWithTags(db *gorm.DB, ts []string) error {
	var tags []*Tag
	for _, t := range ts {
		tags = append(tags, &Tag{Name: t})
	}
	tx := db.Begin()
	onConflict := "ON CONFLICT ON CONSTRAINT (tags_name_key) DO NOTHING"
	for _, t := range tags {
		if err := db.Set("gorm:insert_options", onConflict).Save(t).Error; err != nil {
			tx.Rollback()
			return err
		}
		img.Tags = append(img.Tags, *t)
	}
	if err := db.Save(img).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
