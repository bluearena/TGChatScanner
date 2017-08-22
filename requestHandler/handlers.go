package requestHandler

import (
	"encoding/json"
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
	Err    error          `json:"error"`
	Images []models.Image `json:"images"`
}

type ChatsJSON struct {
	Err   string        `json:"error"`
	Chats []models.Chat `json:"chats"`
}

var user_key = "user"

func getImages(w http.ResponseWriter, req *http.Request) {
	l := appContext.Logger
	values := req.URL.Query()
	img := models.Image{}

	imgs, err := img.GetImgByParams(appContext.Db, values)

	if err != nil {
		l.Println(err)
	}
	response := ImagesJSON{Err: nil,
		Images: imgs}
	responseJSON, err := json.Marshal(response)
	if err == nil {
		writeResponse(w, string(responseJSON), http.StatusTeapot)
		l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusTeapot)
		return
	} else {
		l.Println(err)
		return
	}
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
	l := appContext.Logger
	values := req.URL.Query()

	user_id, err := strconv.ParseUint(values["user_id"][0], 10, 64)

	if err != nil {
		l.Fatal(err)
	}

	user := models.User{TGID: user_id}

	if err := user.GetUsersChats(appContext.Db); err != nil {
		l.Println(err)
		response := ChatsJSON{Err: "system error",
			Chats: nil}
		responseJSON, _ := json.Marshal(response)
		writeResponse(w, string(responseJSON), http.StatusInternalServerError)
		l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusTeapot)
		return
	} else {
		response := ChatsJSON{Err: "",
			Chats: user.Chats}
		responseJSON, _ := json.Marshal(response)
		writeResponse(w, string(responseJSON), http.StatusTeapot)
		l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusTeapot)
		return
	}
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
