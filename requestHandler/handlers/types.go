package handlers

import (
	"github.com/zwirec/TGChatScanner/models"
	"net/http"
	"encoding/json"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
)

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

func (uj *UserJSON) Response(w http.ResponseWriter, r *http.Request, status int){
	responseJSON, err := json.Marshal(uj)
	if err != nil {
		writeResponse(w, nil, http.StatusInternalServerError)
		appContext.ErrLogger.Println(err)
		logHttpRequest(appContext.AccessLogger, r, http.StatusInternalServerError)
		return
	}
	writeResponse(w, string(responseJSON), status)
	logHttpRequest(appContext.AccessLogger, r, status)
}

func (tj *TagsJSON) Response(w http.ResponseWriter, r *http.Request, status int){
	responseJSON, err := json.Marshal(tj)
	if err != nil {
		writeResponse(w, nil, http.StatusInternalServerError)
		appContext.ErrLogger.Println(err)
		logHttpRequest(appContext.AccessLogger, r, http.StatusInternalServerError)
		return
	}
	writeResponse(w, string(responseJSON),status)
	logHttpRequest(appContext.AccessLogger, r, status)
}

func (ij *ImagesJSON) Response(w http.ResponseWriter, r *http.Request, status int){
	responseJSON, err := json.Marshal(ij)
	if err != nil {
		writeResponse(w, nil, http.StatusInternalServerError)
		appContext.ErrLogger.Println(err)
		logHttpRequest(appContext.AccessLogger, r, http.StatusInternalServerError)
		return
	}
	writeResponse(w, string(responseJSON), status)
	logHttpRequest(appContext.AccessLogger, r, status)
}

func (cj *ChatsJSON) Response(w http.ResponseWriter, r *http.Request, status int){
	responseJSON, err := json.Marshal(cj)
	if err != nil {
		writeResponse(w, nil, http.StatusInternalServerError)
		appContext.ErrLogger.Println(err)
		logHttpRequest(appContext.AccessLogger, r, http.StatusInternalServerError)
		return
	}
	writeResponse(w, string(responseJSON), status)
	logHttpRequest(appContext.AccessLogger, r, status)
}