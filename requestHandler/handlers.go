package requestHandler

import (
	"fmt"
	"net/http"
)

type ResponseJSON struct {
	Err   string      `json:"error,omitempty"`
	Model interface{} `json:"entity,omitempty"`
}

// Get images for chat (only for authorized)
func getImages(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "images.get")
	return
}

// Get chats for token (only for authorized)
func getChat(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "chat.get")
	return
}

// Get chats for user (only for authorized)
func getChats(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "chats.get")
	return
}

// Get tags for chat (only for authorized)
func getTags(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "tags.get")
	return
}
