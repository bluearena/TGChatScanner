package requestHandler

import (
	"encoding/json"
	"fmt"
	"github.com/zwirec/TGChatScanner/models"
	"log"
	"net/http"
	"strconv"
)

type UserJSON struct {
	Err   string       `json:"error,omitempty"`
	Model *models.User `json:"entity,omitempty"`
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

var user_key = "user"

func getImages(w http.ResponseWriter, req *http.Request) {
	err_l := appContext.SysLogger
	acc_l := appContext.AccessLogger

	values := req.URL.Query()
	img := models.Image{}

	imgs, err := img.GetImgByParams(appContext.Db, values)

	if err != nil {
		response := ImagesJSON{Err: "server error",
			Images: nil}
		responseJSON, err := json.Marshal(response)

		if err == nil {
			writeResponse(w, string(responseJSON), http.StatusInternalServerError)
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			return
		} else {
			writeResponse(w, nil, http.StatusInternalServerError)
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			err_l.Println(err)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusOK)
		logHttpRequest(acc_l, req, http.StatusOK)
	}

	response := ImagesJSON{Err: "",
		Images: imgs}
	responseJSON, err := json.Marshal(response)

	if err == nil {
		writeResponse(w, string(responseJSON), http.StatusOK)
		logHttpRequest(acc_l, req, http.StatusOK)
		return
	} else {
		err_l.Println(err)
		logHttpRequest(acc_l, req, http.StatusOK)
		return
	}
	return
}

func getChatTags(w http.ResponseWriter, req *http.Request) {
	err_l := appContext.SysLogger
	acc_l := appContext.AccessLogger

	values := req.URL.Query()

	chat_id, err := strconv.ParseInt(values["chat_id"][0], 10, 64)

	if err != nil {
		response := TagsJSON{Err: "invalid chat_id",
			Tags: nil}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			err_l.Println(err)
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusBadRequest)
		logHttpRequest(acc_l, req, http.StatusBadRequest)
		return
	}

	chat := models.Chat{TGID: chat_id}

	tags, err := chat.GetTags(appContext.Db)

	if err != nil {
		err_l.Println(err)
		response := TagsJSON{Err: "system error",
			Tags: nil}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			err_l.Println(err)
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusBadRequest)

		logHttpRequest(acc_l, req, http.StatusBadRequest)
		return
	} else {
		response := TagsJSON{Err: "",
			Tags: tags}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			err_l.Println(err)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusOK)
		logHttpRequest(acc_l, req, http.StatusOK)
		return
	}

}

func getUserTags(w http.ResponseWriter, req *http.Request) {
	err_l := appContext.SysLogger
	acc_l := appContext.AccessLogger

	values := req.URL.Query()

	user_id, err := strconv.ParseInt(values["user_id"][0], 10, 32)

	if err != nil {
		response := TagsJSON{Err: "invalid chat_id",
			Tags: nil}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			err_l.Println(err)
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusBadRequest)
		logHttpRequest(acc_l, req, http.StatusBadRequest)
		return
	}

	user := models.User{TGID: int(user_id)}

	tags, err := user.GetTags(appContext.Db)

	if err != nil {
		err_l.Println(err)
		response := TagsJSON{Err: "system error",
			Tags: nil}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			err_l.Println(err)
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusBadRequest)

		logHttpRequest(acc_l, req, http.StatusBadRequest)
		return
	} else {
		response := TagsJSON{Err: "",
			Tags: tags}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			err_l.Println(err)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusOK)
		logHttpRequest(acc_l, req, http.StatusOK)
		return
	}

}

func getChats(w http.ResponseWriter, req *http.Request) {
	err_l := appContext.SysLogger
	acc_l := appContext.AccessLogger

	user := req.Context().Value(user_key).(models.User)

	if err := user.GetUsersChats(appContext.Db); err != nil {
		err_l.Println(err)
		response := ChatsJSON{Err: "system error",
			Chats: nil}
		responseJSON, _ := json.Marshal(response)
		writeResponse(w, string(responseJSON), http.StatusInternalServerError)

		logHttpRequest(acc_l, req, http.StatusInternalServerError)
		return
	} else {
		response := ChatsJSON{Err: "",
			Chats: user.Chats}
		responseJSON, _ := json.Marshal(response)
		writeResponse(w, string(responseJSON), http.StatusOK)
		logHttpRequest(acc_l, req, http.StatusOK)
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

func logHttpRequest(l *log.Logger, req *http.Request, code int) {
	l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, code)
}
