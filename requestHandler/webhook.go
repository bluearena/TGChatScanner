package requestHandler

import (
	"bytes"
	"errors"
	"github.com/rs/xid"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/models"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	UserStatsUrl     = "/stats"
	MaxFailedUpdates = 100
)

var (
	ErrUnexpectedCommand = errors.New("unexpected command")
)

func BotUpdateHanlder(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	message := ctx.Value(messageKey).(*TGBotApi.Message)

	accLog := appContext.AccessLogger
	errLog := appContext.SysLogger

	localCtx := make(map[string]interface{})
	localCtx["from"] = message.Chat.Id

	if message.Document != nil && isPicture(message.Document.MimeType) {
		fb := &FileBasic{
			FileId:  message.Document.FileId,
			Type:    "photo",
			Sent:    time.Unix(int64(message.Date), 0),
			Context: localCtx,
		}
		appContext.DownloadRequests <- fb
	}
	if pl := len(message.Photo); pl != 0 {
		photo := message.Photo[pl-1]
		fb := &FileBasic{
			FileId:  photo.FileId,
			Type:    "photo",
			Sent:    time.Unix(int64(message.Date), 0),
			Context: localCtx,
		}
		appContext.DownloadRequests <- fb
	} else if len(message.Entities) != 0 && message.Entities[0].Type == "bot_command" {
		if err := BotCommandRouter(message); err != nil {
			errLog.Println(err)
			if err == ErrUnexpectedCommand {
				logHttpRequest(accLog, req, http.StatusOK)
				w.WriteHeader(http.StatusOK)
			} else {
				logHttpRequest(accLog, req, http.StatusInternalServerError)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
	logHttpRequest(accLog, req, http.StatusOK)
	w.WriteHeader(http.StatusOK)
}

func BotCommandRouter(message *TGBotApi.Message) error {
	r := regexp.MustCompile(`/(start(?:group)?|mystats|wantscan)?\s*`)
	command := r.FindStringSubmatch(message.Text)
	if len(command) == 0 {
		return ErrUnexpectedCommand
	}
	switch command[1] {
	case "start":
		fallthrough
	case "startgroup":
		ch := models.Chat{
			TGID:  message.Chat.Id,
			Title: message.Chat.Title,
		}
		err := ch.CreateIfNotExists(appContext.Db)
		if err != nil {
			return err
		}
		hello := "Hello, chat " + message.Chat.Title

		_, err = appContext.BotApi.SendMessage(message.Chat.Id, hello, true)
		return err
	case "wantscan":
		err := AddSubscription(&message.From, &message.Chat)
		return err
	case "mystats":
		usr := models.User{
			TGID:     message.From.Id,
			Username: message.From.UserName,
		}

		err := usr.CreateIfNotExists(appContext.Db)
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
func AddSubscription(user *TGBotApi.User, chat *TGBotApi.Chat) (err error) {
	var username string
	if user.UserName != "" {
		username = user.UserName
	} else {
		username = user.FirstName
	}
	db := appContext.Db

	u := &models.User{
		TGID:     user.Id,
		Username: username,
	}

	ch := &models.Chat{
		TGID:  chat.Id,
		Title: chat.Title,
	}
	tx := db.Begin()

	err = ch.CreateIfNotExists(db)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = u.CreateIfNotExists(db)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = db.Model(u).Association("Chats").Append([]models.Chat{*ch}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func SetUserToken(userId int) (string, error) {
	guid := xid.New()
	t := &models.Token{
		Token:     guid.String(),
		ExpiredTo: time.Now().AddDate(0, 0, 1),
		UserID:    userId,
	}

	if err := appContext.Db.Save(t).Error; err != nil {
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
