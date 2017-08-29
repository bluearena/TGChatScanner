package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/models"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func MiddlewareLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			token := req.Header.Get("X-User-Token")

			if token == "" {
				response := UserJSON{
					Err:  "X-User-Token isn't set",
					User: nil,
				}
				response.Response(w, req, http.StatusForbidden)
				return
			}

			memcache := appContext.Cache

			var user *models.User

			u, expire, ok := memcache.GetWithExpiration(token)

			if ok {
				user = u.(*models.User)
			}

			if !ok || (ok && expire.Before(time.Now().Add(2*time.Minute))) {
				tok := models.Token{Token: token}
				expired_to := tok.ExpiredTo
				user = tok.GetUserByToken(appContext.DB)
				if expired_to.Add(time.Minute).Before(time.Now()) {
					memcache.Set(token, user, time.Minute)
				}
			}

			if user == nil {
				response := UserJSON{
					Err:  "incorrect user_id or tokens lifetime is expired",
					User: nil,
				}
				response.Response(w, req, http.StatusForbidden)

			} else {
				ctx := context.WithValue(req.Context(), UserKey, user)
				next.ServeHTTP(w, req.WithContext(ctx))
			}

		} else {
			response := UserJSON{
				Err:  "method not allowed",
				User: nil,
			}
			response.Response(w, req, http.StatusMethodNotAllowed)
		}

	})
}

func BotRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		update := ctx.Value(UpdateKey).(*TGBotAPI.Update)
		message := extractMessage(update)
		if message.From.Id == 129039602{
			responseAndLog(w, req, http.StatusOK)
			return
		}
		var upType string
		if len(message.Entities) != 0 && message.Entities[0].Type == "bot_command" {
			upType = CommandType
		} else if message.Document != nil && isPicture(message.Document.MimeType) {
			upType = DocumentType
		} else if l := len(message.Photo); l != 0 {
			upType = PictureType
		} else {
			responseAndLog(w, req, http.StatusOK)
			return
		}
		ctx = context.WithValue(ctx, UpdateTypeKey, upType)
		ctx = context.WithValue(ctx, MessageKey, message)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func ExtractUpdate(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		errLog := appContext.ErrLogger

		if err != nil {
			errLog.Printf("error during reading bot request: %s", err)
			responseAndLog(w, req, http.StatusOK)
			return
		}

		var update TGBotAPI.Update
		err = json.Unmarshal(body, &update)
		if err != nil {
			errLog.Printf("error during unmarshaling request: %s: %s", req.URL.String(), err)
			responseAndLog(w, req, http.StatusInternalServerError)
			return
		}
		if getUpdatesCount(&update) > MaxFailedUpdates {
			errLog.Printf("max failed updates number exceeded on %d", update.UpdateId)
			responseAndLog(w, req, http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(req.Context(), UpdateKey, &update)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func extractMessage(update *TGBotAPI.Update) (message *TGBotAPI.Message) {
	if update.Message != nil {
		message = update.Message
	} else if update.EditedMessage != nil {
		message = update.EditedMessage
	}
	if message.Chat.Username != "" {
		message.Chat.Title = message.Chat.Username
	}
	title := message.Chat.Title
	if title == "" {
		title = message.Chat.Username
	}
	return message
}

func getUpdatesCount(update *TGBotAPI.Update) int {
	updateID := update.UpdateId
	updateIDKey := strconv.Itoa(updateID)
	updatesCount, err := appContext.Cache.IncrementInt(updateIDKey, 1)

	if err != nil {
		appContext.Cache.Set(updateIDKey, 1, time.Minute)
		updatesCount = 1
	}
	return updatesCount
}

func responseAndLog(w http.ResponseWriter, req *http.Request, status int) {
	logHttpRequest(appContext.AccessLogger, req, status)
	w.WriteHeader(status)
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
