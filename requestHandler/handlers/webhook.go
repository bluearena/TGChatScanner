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
	CommandsMatcher      *regexp.Regexp
)

func init() {
	rp := `^/((?:no|want)scan|start(?:\s+newtoken)?|mystats|startgroup)\s*(@chatscannerbot)?$`
	CommandsMatcher = regexp.MustCompile(rp)
}

func BotUpdateHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	updateID := ctx.Value(UpdateKey).(*TGBotAPI.Update).UpdateId
	message := ctx.Value(MessageKey).(*TGBotAPI.Message)
	uptype := ctx.Value(UpdateTypeKey).(string)

	accLog := appContext.AccessLogger
	errLog := appContext.ErrLogger

	err := autoCreateChat(message)
	if err != nil {
		responseAndLog(w, req, http.StatusInternalServerError)
		return
	}

	status := http.StatusOK

	switch uptype {
	case CommandType:
		err = BotCommandRouter(message)
		if err == ErrUnexpectedCommand {
			errLog.Printf("Unexpected command: %s", message.Text)
		} else if err != nil {
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
	}
	responseAndLog(w, req, status)
}

func BotCommandRouter(message *TGBotAPI.Message) error {
	command := CommandsMatcher.FindStringSubmatch(message.Text)
	if len(command) < 2 {
		return ErrUnexpectedCommand
	}
	chatId := message.Chat.Id
	switch command[1] {
	case "start":
		fallthrough
	case "startgroup":
		hello := createHello(&message.Chat)
		_, err := appContext.BotAPI.SendMessage(message.Chat.Id, hello, true)
		return err
	case "start newtoken":
		fallthrough
	case "mystats":
		if err := authorizeAccess(message); err != nil {
			appContext.ErrLogger.Printf("Fail to authorize %v", message.From)
			return responseTryAgain(chatId)
		}
	case "wantscan":
		err := AddSubscription(&message.From, &message.Chat)
		if err != nil {
			appContext.ErrLogger.Printf("Fail to add subscription %v", message.From)
			return responseTryAgain(chatId)
		}
		answer := "Subscription +"
		_, err = appContext.BotAPI.SendMessage(chatId, answer, true)
		return err
	case "noscan":
		if err := removeSubsription(message.From.Id, chatId); err != nil {
			return responseTryAgain(chatId)
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
	if err := u.CreateIfNotExists(tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&u).Association("Chats").Append(ch).Error; err != nil {
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

func createHello(chat *TGBotAPI.Chat) string {
	var hello bytes.Buffer
	hello.WriteString("Hello, ")
	if chat.Type == "private" {
		hello.WriteString("user ")
	} else {
		hello.WriteString("chat ")
	}
	hello.WriteString(chat.Title)
	hello.WriteString("\nAvailable commands:\n")
	hello.WriteString("/wantscan to add chat to your subscriptions\n")
	hello.WriteString("/mystats to get link on your image feed\n")
	hello.WriteString("/noscan to remove chat from your subscriptions\n")
	return hello.String()
}

func authorizeAccess(message *TGBotAPI.Message) error {
	if message.Chat.Type != "private" {
		botUrl := "https://telegram.me/chatscannerbot?start"
		answer := "Ask here: " + botUrl
		_, err := appContext.BotAPI.SendMessage(message.Chat.Id, answer, true)
		return err
	}
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
	return nil
}

func removeSubsription(userId int, chatId int64) error {
	usr := models.User{
		TGID: userId,
	}
	ch := models.Chat{
		TGID: chatId,
	}
	db := appContext.DB
	err := db.Model(&usr).Association("Chats").Delete(&ch).Error
	if err != nil {
		appContext.ErrLogger.Printf("fail on removing user-chat: user %v, chat %v: %s", userId, chatId, err)
	}
	return err
}

func responseTryAgain(chatId int64) error {
	_, err := appContext.BotAPI.SendMessage(chatId, "Try again", true)
	return err
}
