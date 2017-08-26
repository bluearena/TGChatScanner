package handlers

import (
	"github.com/zwirec/TGChatScanner/models"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	"log"
	"net/http"
	"strconv"
)

func GetImages(w http.ResponseWriter, req *http.Request) {
	errLog := appContext.ErrLogger
	values := req.URL.Query()
	img := models.Image{}

	user := req.Context().Value(UserKey).(*models.User)

	if len((*user).Chats) == 0 {
		response := ImagesJSON{
			Err:    "",
			Images: nil,
		}
		response.Response(w, req, http.StatusOK)
		return
	}

	imgs, err := img.GetImgByParams(appContext.DB, values, user)

	if err != nil {
		response := ImagesJSON{
			Err:    "server error",
			Images: nil,
		}
		errLog.Println(err)
		response.Response(w, req, http.StatusInternalServerError)
		return
	}

	response := ImagesJSON{
		Err:    "",
		Images: imgs,
	}
	response.Response(w, req, http.StatusOK)
	return
}

func GetChatTags(w http.ResponseWriter, req *http.Request) {
	errLog := appContext.ErrLogger

	values := req.URL.Query()

	chatid, ok := values["chat_id"]

	if !ok {
		response := TagsJSON{
			Err:  "invalid chat_id",
			Tags: nil,
		}
		response.Response(w, req, http.StatusBadRequest)
		return
	}

	chat_id, err := strconv.ParseInt(chatid[0], 10, 64)

	if err != nil {
		response := TagsJSON{
			Err:  "invalid chat_id",
			Tags: nil,
		}
		errLog.Println(err)
		response.Response(w, req, http.StatusBadRequest)
		return
	}

	chat := models.Chat{TGID: chat_id}

	user := req.Context().Value(UserKey).(*models.User)

	find := false

	for _, ch := range user.Chats {
		if chat.TGID == ch.TGID {
			find = true
			break
		}
	}

	if !find {
		response := TagsJSON{
			Err:  "user wasn't subscribed on this chat",
			Tags: nil,
		}
		response.Response(w, req, http.StatusNotFound)
		return
	}

	tags, err := chat.GetTags(appContext.DB)

	if err != nil {
		response := TagsJSON{
			Err:  "system error",
			Tags: nil,
		}
		errLog.Println(err)
		response.Response(w, req, http.StatusBadRequest)
		return
	} else {
		response := TagsJSON{
			Err:  "",
			Tags: tags,
		}
		response.Response(w, req, http.StatusOK)
		return
	}

}

func GetUserTags(w http.ResponseWriter, req *http.Request) {
	errLog := appContext.ErrLogger

	user := req.Context().Value(UserKey).(*models.User)

	tags, err := user.GetTags(appContext.DB)

	if err != nil {
		response := TagsJSON{
			Err:  "system error",
			Tags: nil,
		}
		errLog.Println(err)
		response.Response(w, req, http.StatusBadRequest)
		return
	} else {
		response := TagsJSON{
			Err:  "",
			Tags: tags,
		}
		response.Response(w, req, http.StatusOK)
		return
	}

}

func GetChats(w http.ResponseWriter, req *http.Request) {
	errLog := appContext.ErrLogger

	user := req.Context().Value(UserKey).(*models.User)

	if err := user.GetUsersChats(appContext.DB); err != nil {
		errLog.Println(err)
		response := ChatsJSON{
			Err:   "system error",
			Chats: nil,
		}
		response.Response(w, req, http.StatusInternalServerError)
		return
	} else {
		response := ChatsJSON{
			Err:   "",
			Chats: user.Chats,
		}

		response.Response(w, req, http.StatusOK)
		return
	}
}

func logHttpRequest(l *log.Logger, req *http.Request, code int) {
	l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, code)
}
