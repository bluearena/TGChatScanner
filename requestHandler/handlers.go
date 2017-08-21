package requestHandler

import (
	"fmt"
	"github.com/zwirec/TGChatScanner/models"
	"net/http"
	"net/url"
	"strconv"
)

type UserJSON struct {
	Err   string       `json:"error,omitempty"`
	Model *models.User `json:"entity,omitempty"`
}

type ImagesJSON struct {
}

var user_key = "user"

func getImages(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()

	imgs := &models.Image{}.GetImgByParams(appContext.Db, values)
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

func validateGETParams(values url.Values) (map[string]interface{}, bool) {
	if values["user_id"] == nil || values["chat_id"] == nil || values["token"] == nil {
		return nil, false
	}
	result := map[string]interface{}{}

	var err error
	result["user_id"], err = strconv.ParseUint(values["user_id"][0], 10, 64)
	result["chat_id"], err = strconv.ParseUint(values["client_id"][0], 10, 64)

	if err != nil {
		return nil, false
	}
	result["token"] = values["token"]
	return result, false
}

func validateLoginParams(values map[string]interface{}) (ok bool) {
	if values["username"] == nil || values["password"] == nil {
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
