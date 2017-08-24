package handlers

import "github.com/zwirec/TGChatScanner/models"

type UserJSON struct {
	Err  string       `json:"error,omitempty"`
	User *models.User `json:"entity,omitempty"`
}

type ImagesJSON struct {
	Err    string         `json:"error"`
	Images []models.Image `json:"images"`
}

type ChatsJSON struct {
	Err   string        `json:"error"`
	Chats []models.Chat `json:"chats"`
}

type TagsJSON struct {
	Err  string       `json:"error"`
	Tags []models.Tag `json:"tags"`
}
