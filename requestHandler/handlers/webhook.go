package handlers

import (
	"bytes"
	"context"
	"errors"
	"github.com/rs/xid"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/models"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	UserStatsURL     = "/stats"
	MaxFailedUpdates = 100
	UpdateTimeout    = 3 * time.Second

	CommandType  = "command"
	PictureType  = "picture"
	DocumentType = "doc"
)

var (
	ErrUnexpectedCommand = errors.New("unexpected command")
	ErrRequestTimeout    = errors.New("request timeout")
)

func BotUpdateHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	updateID := ctx.Value(UpdateKey).(*TGBotAPI.Update).UpdateId
	message := ctx.Value(MessageKey).(*TGBotAPI.Message)
	uptype := ctx.Value(UpdateTypeKey).(string)

	accLog := appContext.AccessLogger
	errLog := appContext.ErrLogger

	err := autoCreateChat(message)
	if err != nil {
		logHttpRequest(accLog, req, http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	status := http.StatusOK

	switch uptype {
	case CommandType:
		err = BotCommandRouter(message)
		if err != nil {
			status = http.StatusInternalServerError
			errLog.Println(err)
		}
		status = http.StatusOK

	case DocumentType:
		fb := file.NewFileBasic(message, "photo", message.Document.FileId)
		ctx, cancel := context.WithTimeout(context.Background(), UpdateTimeout)
		defer cancel()
		err = handleFile(fb, ctx)

	case PictureType:
		photo := message.Photo[len(message.Photo)-1]
		fb := file.NewFileBasic(message, "photo", photo.FileId)
		ctx, cancel := context.WithTimeout(context.Background(), UpdateTimeout)
		defer cancel()
		err = handleFile(fb, ctx)
	}

	if err != nil {
		status = http.StatusInternalServerError
		errLog.Printf("update %d: %s", updateID, err)
		return
	}
	logHttpRequest(accLog, req, status)
	w.WriteHeader(status)
}

func BotCommandRouter(message *TGBotAPI.Message) error {
	r := regexp.MustCompile(`/(start(?:group)?|mystats|wantscan)?\s*`)
	command := r.FindStringSubmatch(message.Text)
	if len(command) == 0 {
		return ErrUnexpectedCommand
	}
	switch command[1] {
	case "start":
		fallthrough
	case "startgroup":
		hello := "Hello, chat " + message.Chat.Title
		_, err := appContext.BotAPI.SendMessage(message.Chat.Id, hello, true)
		return err
	case "wantscan":
		err := AddSubscription(&message.From, &message.Chat)
		return err
	case "mystats":
		usr := models.User{
			TGID:     message.From.Id,
			Username: message.From.UserName,
		}
		err := usr.CreateIfNotExists(appContext.DB)
		token, err := SetUserToken(message.From.Id)
		if err != nil {
			return err
		}
		us := BuildUserStatURL(token)
		_, err = appContext.BotAPI.SendMessage(message.Chat.Id, us, true)
		if err != nil {
			return err
		}
	}
	return nil
}
func AddSubscription(user *TGBotAPI.User, chat *TGBotAPI.Chat) (err error) {
	var username string
	if user.UserName != "" {
		username = user.UserName
	} else {
		username = user.FirstName
	}
	db := appContext.DB

	ch := &models.Chat{
		TGID:  chat.Id,
		Title: chat.Title,
	}
	u := &models.User{
		TGID:     user.Id,
		Username: username,
	}
	tx := db.Begin()
	if err := tx.Model(&u).Association("Chats").Append(ch).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := u.CreateIfNotExists(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func SetUserToken(userId int) (string, error) {
	guid := xid.New()
	t := &models.Token{
		Token:     guid.String(),
		ExpiredTo: time.Now().AddDate(0, 0, 1),
		UserID:    userId,
	}

	if err := appContext.DB.Save(t).Error; err != nil {
		return "", err
	}
	return t.Token, nil
}

func BuildUserStatURL(token string) string {
	var buff bytes.Buffer
	buff.WriteString(appContext.Hostname)
	buff.WriteString(UserStatsURL)
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

func handleFile(fb *file.FileBasic, ctx context.Context) error {

	fb.BasicContext = ctx
	appContext.DownloadRequests <- fb

	select {
	case <-fb.BasicContext.Done():
		return ErrRequestTimeout
	case err := <-fb.Errorc:
		if err != nil {
			return err
		}
	}
	return nil
}

func autoCreateChat(message *TGBotAPI.Message) error {

	title := message.Chat.Title
	if title == "" {
		title = message.Chat.Username
	}

	chat := &models.Chat{
		TGID:  message.Chat.Id,
		Title: title,
	}
	return chat.CreateIfNotExists(appContext.DB)
}
