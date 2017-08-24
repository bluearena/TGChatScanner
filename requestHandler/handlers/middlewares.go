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
		err_l := appContext.ErrLogger
		acc_l := appContext.AccessLogger

		if req.Method == "GET" {
			token := req.Header.Get("X-User-Token")

			if token == "" {
				response := UserJSON{Err: "X-User-Token isn't set",
					User: nil}

				responseJSON, err := json.Marshal(response)
				if err == nil {
					writeResponse(w, string(responseJSON), http.StatusForbidden)
					acc_l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusForbidden)
					return
				} else {
					err_l.Println(err)
					acc_l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusInternalServerError)
					return
				}
			}

			memcache := appContext.Cache

			user, expire, _ := memcache.GetWithExpiration(token)

			if user == nil || expire.Before(time.Now().Add(2*time.Minute)) {
				tok := models.Token{Token: token}
				expired_to := tok.ExpiredTo
				user = tok.GetUserByToken(appContext.DB)
				if expired_to.Add(time.Minute).Before(time.Now()) {
					memcache.Set(token, user, time.Minute)
				}

			}

			if user == nil {
				response := UserJSON{Err: "incorrect user_id or tokens lifetime is expired",
					User: nil}
				responseJSON, err := json.Marshal(response)
				if err == nil {
					writeResponse(w, string(responseJSON), http.StatusTeapot)
					acc_l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusTeapot)
					return
				} else {
					err_l.Println(err)
					acc_l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusInternalServerError)
					return
				}
			} else {
				ctx := context.WithValue(req.Context(), UserKey, user)
				next.ServeHTTP(w, req.WithContext(ctx))
			}

		} else {
			response := UserJSON{Err: "method not allowed",
				User: nil}
			responseJSON, err := json.Marshal(response)
			if err == nil {
				writeResponse(w, string(responseJSON), http.StatusMethodNotAllowed)
				acc_l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusMethodNotAllowed)
				return
			} else {
				acc_l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusInternalServerError)
				err_l.Println(err)
				return
			}
		}

	})
}

func ChatAutoStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		acc_l := appContext.AccessLogger
		errLog := appContext.ErrLogger
		if err != nil {
			errLog.Printf("error during reading bot request: %s", err)
			logHttpRequest(acc_l, req, http.StatusOK)
			w.WriteHeader(http.StatusOK)
			return
		}

		var update TGBotApi.Update
		err = json.Unmarshal(body, &update)
		if err != nil {
			errLog.Printf("error during unmarshaling request: %s: %s", req.URL.String(), err)
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		upid := update.UpdateId
		upidKey := strconv.Itoa(upid)
		up_num, err := appContext.Cache.IncrementInt(upidKey, 1)

		if err != nil {
			appContext.Cache.Set(upidKey, 1, time.Minute)
			up_num = 1
		}
		if up_num > MaxFailedUpdates {
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			errLog.Printf("Max failed updates number exceeded on %d", upid)
			w.WriteHeader(http.StatusOK)
			return
		}

		var message *TGBotApi.Message

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

		chat := &models.Chat{
			TGID:  message.Chat.Id,
			Title: title,
		}
		err = chat.CreateIfNotExists(appContext.DB)
		if err != nil {
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(req.Context(), appContext.MessageKey, message)

		next.ServeHTTP(w, req.WithContext(ctx))
		return
	})
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
