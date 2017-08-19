package requestHandler

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"log"
	"encoding/json"
	"github.com/zwirec/TGChatScanner/models"
)

func registerUser(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)

	var values map[string]interface{}

	if err := decoder.Decode(&values); err != nil {
		writeResponse(w, "Incorrect JSON\n", http.StatusBadRequest)
		return
	}

	if !validateRegParam(values) {
		writeResponse(w, "Incorrect params\n", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(values["password"].(string)), bcrypt.DefaultCost)

	if err != nil {
		writeResponse(w, nil, http.StatusInternalServerError)
	}

	logger := req.Context().Value(loggerContextKey).(*log.Logger)

	user := models.User{Username: values["username"].(string),
		Password: string(hash),
		Email: values["email"].(string)}

	_, err = user.Register(appContext.Db)

	if err != nil {
		logger.Println(err)
		writeResponse(w, nil, http.StatusInternalServerError)
		return
	}
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

func validateRegParam(values map[string]interface{}) (ok bool) {
	if values["username"] == nil || values["password"] == nil || values["email"] == nil {
		return false
	}
	return true
}

func writeResponse(w http.ResponseWriter, data interface{}, status int) error {
	w.WriteHeader(status)
	if data != nil {
		_, err := fmt.Fprint(w, data)

		if err != nil {
			return err
		}
	}
	return nil
}
