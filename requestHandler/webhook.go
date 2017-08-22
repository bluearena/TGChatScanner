package requestHandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/rs/xid"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/models"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	UserStatsUrl = "/stats"
)

var (
	ErrUnexpectedCommand = errors.New("unexpected command")
)

func BotUpdateHanlder(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	acc_l := req.Context().Value(accLoggerKey).(*log.Logger)
	sys_l := req.Context().Value(sysLoggerKey).(*log.Logger)
	if err != nil {
		sys_l.Printf("Error during reading bot request: %s", err)
		logHttpRequest(acc_l, req, http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var update TGBotApi.Update
	err = json.Unmarshal(body, &update)
	if err != nil {
		sys_l.Printf("Error during unmarshaling request: %s", req.URL.String(), err)
		logHttpRequest(acc_l, req, http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message *TGBotApi.Message

	if update.Message != nil {
		message = update.Message
	} else if update.EditedMessage != nil {
		message = update.EditedMessage
	}

	ctx := make(map[string]interface{})
	ctx["from"] = message.Chat.Id

	if message.Document != nil && isPicture(message.Document.MimeType) {
		fb := &FileBasic{
			FileId:  message.Document.FileId,
			Type:    "photo",
			Sent:    time.Unix(int64(message.Date), 0),
			Context: ctx,
		}
		appContext.DownloadRequests <- fb
	}
	if pl := len(message.Photo); pl != 0 {
		photo := message.Photo[pl-1]
		fb := &FileBasic{
			FileId:  photo.FileId,
			Type:    "photo",
			Sent:    time.Unix(int64(message.Date), 0),
			Context: ctx,
		}
		appContext.DownloadRequests <- fb
	} else if len(message.Entities) != 0 && message.Entities[0].Type == "bot_command" {
		if err := BotCommandRouter(message); err != nil {
			sys_l.Printf("Command: %s, error: %s", err)
			if err == ErrUnexpectedCommand {
				logHttpRequest(acc_l, req, http.StatusOK)
				w.WriteHeader(http.StatusOK)
			} else {
				logHttpRequest(acc_l, req, http.StatusInternalServerError)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func BotCommandRouter(message *TGBotApi.Message) error {
	r := regexp.MustCompile(`\/(start(?:group)?|mystats)?\s*`)
	command := r.FindStringSubmatch(message.Text)
	if len(command) == 0 {
		return ErrUnexpectedCommand
	}
	switch command[1] {
	case "start":
	case "startgroup":

	case "wantscan":
		err := AddSubsription(&message.From, &message.Chat)
		return err
	case "mystats":
		token, err := SetUserToken(message.From.Id)
		if err != nil {
			return err
		}

		us := BuildUserStatUrl(token)
		_, err = appContext.BotApi.SendMessage(message.Chat.Id, us, true)
		if err != nil {
			return err
		}
	}
	return nil
}
func AddSubsription(user *TGBotApi.User, chat *TGBotApi.Chat) error {
	var username string
	if user.UserName != "" {
		username = user.UserName
	} else {
		username = user.FirstName
	}
	db := appContext.Db

	ch := models.Chat{
		TGID:  chat.Id,
		Title: chat.Title,
	}
	u := &models.User{
		TGID:     user.Id,
		Username: username,
		Chats:    []models.Chat{ch},
	}

	return db.Save(u).Error
}

func SetUserToken(userId int) (string, error) {
	guid := xid.New()
	t := &models.Token{
		Token:     guid.String(),
		ExpiredTo: time.Now().AddDate(0, 0, 1),
		UserID:    userId,
	}

	if err := appContext.Db.Create(t).Error; err != nil {
		return "", err
	}
	return t.Token, nil
}

func BuildUserStatUrl(token string) string {
	var buff bytes.Buffer
	buff.WriteString(appContext.Hostname)
	buff.WriteString(UserStatsUrl)
	buff.WriteString("?")
	params := url.Values{}
	params.Add("token", token)
	buff.WriteString(params.Encode())
	return buff.String()
}

func isPicture(mtype string) bool {
	m, _, err := mime.ParseMediaType(mtype)

	if err != nil {
		return false
	}
	if strings.HasPrefix(m, "image") {
		return true
	}
	return false
}
