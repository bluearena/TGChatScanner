package requestHandler

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func registerUser(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	data := req.PostForm
	hash, err := bcrypt.GenerateFromPassword([]byte(data["password"][0]), bcrypt.DefaultCost)
	logger := req.Context().Value(loggerContextKey).(AppContext).Logger

	if err != nil {
		logger.Println(err)
	}

	fmt.Fprintf(w, "Hash to store: %s", string(hash))
	return
}

func loginUser(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "users.login")
}

func getImages(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "images.get")
	return
}

func restoreImages(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "images.restore")
	return
}
func removeImages(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "images.remove")
	return
}
func getChats(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "chats.get")
	return
}
func getTags(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "tags.get")
	return
}
func removeSubs(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "subs.remove")
	return
}
