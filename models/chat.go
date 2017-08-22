package models

import "github.com/jinzhu/gorm"

type Chat struct {
	gorm.Model
	TGID  uint64 `gorm:"primary_key:true" json:"chat_id"`
	Title string `json:"title"`
	Users []User `json:"-"`
	//Image uint
	Images []Image `gorm:"ForeignKey:ChatID;AssociationForeignKey:TGID" json:"images,omitempty"`
	Tags   []Tag   `gorm:"many2many:chats_tags;AssociationForeignKey:TGID;ForeignKey:ChatID"`
}

func (ch *Chat) GetTags(db *gorm.DB) error {
	db.Model(&Chat{}).
		Preload("Tags").
		Where("tg_id = ?", ch.TGID).
		Find(ch)
	return db.Error
}
