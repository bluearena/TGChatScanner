package handlers

import (
	"encoding/json"
	"github.com/zwirec/TGChatScanner/models"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	"log"
	"net/http"
	"strconv"
)

var UserKey = "user"

func GetImages(w http.ResponseWriter, req *http.Request) {
	errLog := appContext.ErrLogger
	accLog := appContext.AccessLogger

	values := req.URL.Query()
	img := models.Image{}

	imgs, err := img.GetImgByParams(appContext.DB, values)

	if err != nil {
		response := ImagesJSON{Err: "server error",
			Images: nil}
		responseJSON, err := json.Marshal(response)

		if err == nil {
			writeResponse(w, string(responseJSON), http.StatusInternalServerError)
			logHttpRequest(accLog, req, http.StatusInternalServerError)
			return
		} else {
			writeResponse(w, nil, http.StatusInternalServerError)
			logHttpRequest(accLog, req, http.StatusInternalServerError)
			errLog.Println(err)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusOK)
		logHttpRequest(accLog, req, http.StatusOK)
	}

	response := ImagesJSON{Err: "",
		Images: imgs}
	responseJSON, err := json.Marshal(response)

	if err == nil {
		writeResponse(w, string(responseJSON), http.StatusOK)
		logHttpRequest(accLog, req, http.StatusOK)
		return
	} else {
		errLog.Println(err)
		logHttpRequest(accLog, req, http.StatusOK)
		return
	}
	return
}

func GetChatTags(w http.ResponseWriter, req *http.Request) {
	errLog := appContext.ErrLogger
	accLog := appContext.AccessLogger

	values := req.URL.Query()

	chatid, ok := values["chat_id"]

	if !ok {
		response := TagsJSON{Err: "invalid chat_id",
			Tags: nil}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			errLog.Println(err)
			logHttpRequest(accLog, req, http.StatusInternalServerError)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusBadRequest)
		logHttpRequest(accLog, req, http.StatusBadRequest)
		return
	}

	chat_id, err := strconv.ParseInt(chatid[0], 10, 64)

	if err != nil {
		response := TagsJSON{Err: "invalid chat_id",
			Tags: nil}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			errLog.Println(err)
			logHttpRequest(accLog, req, http.StatusInternalServerError)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusBadRequest)
		logHttpRequest(accLog, req, http.StatusBadRequest)
		return
	}

	chat := models.Chat{TGID: chat_id}

	tags, err := chat.GetTags(appContext.DB)

	if err != nil {
		errLog.Println(err)
		response := TagsJSON{Err: "system error",
			Tags: nil}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			errLog.Println(err)
			logHttpRequest(accLog, req, http.StatusInternalServerError)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusBadRequest)
		logHttpRequest(accLog, req, http.StatusBadRequest)
		return
	} else {
		response := TagsJSON{Err: "",
			Tags: tags}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			logHttpRequest(accLog, req, http.StatusInternalServerError)
			errLog.Println(err)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusOK)
		logHttpRequest(accLog, req, http.StatusOK)
		return
	}

}

func GetUserTags(w http.ResponseWriter, req *http.Request) {
	errLog := appContext.ErrLogger
	accLog := appContext.AccessLogger

	user := req.Context().Value(UserKey).(*models.User)

	tags, err := user.GetTags(appContext.DB)

	if err != nil {
		errLog.Println(err)
		response := TagsJSON{Err: "system error",
			Tags: nil}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			errLog.Println(err)
			logHttpRequest(accLog, req, http.StatusInternalServerError)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusBadRequest)

		logHttpRequest(accLog, req, http.StatusBadRequest)
		return
	} else {
		response := TagsJSON{Err: "",
			Tags: tags}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			writeResponse(w, nil, http.StatusInternalServerError)
			logHttpRequest(accLog, req, http.StatusInternalServerError)
			errLog.Println(err)
			return
		}
		writeResponse(w, string(responseJSON), http.StatusOK)
		logHttpRequest(accLog, req, http.StatusOK)
		return
	}

}

func GetChats(w http.ResponseWriter, req *http.Request) {
	errLog := appContext.ErrLogger
	accLog := appContext.AccessLogger

	user := req.Context().Value(UserKey).(*models.User)

	if err := user.GetUsersChats(appContext.DB); err != nil {
		errLog.Println(err)
		response := ChatsJSON{Err: "system error",
			Chats: nil}
		responseJSON, _ := json.Marshal(response)
		writeResponse(w, string(responseJSON), http.StatusInternalServerError)

		logHttpRequest(accLog, req, http.StatusInternalServerError)
		return
	} else {
		response := ChatsJSON{Err: "",
			Chats: user.Chats}
		responseJSON, _ := json.Marshal(response)
		writeResponse(w, string(responseJSON), http.StatusOK)
		logHttpRequest(accLog, req, http.StatusOK)
		return
	}
}

func logHttpRequest(l *log.Logger, req *http.Request, code int) {
	l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, code)
}
