package requestHandler

import (
	"context"
	"encoding/json"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/models"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func middlewareLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err_l := appContext.SysLogger
		acc_l := appContext.AccessLogger

		if req.Method == "GET" {
			token := req.Header.Get("X-User-Token")

			if token == "" {
				response := UserJSON{Err: "X-User-Token isn't set",
					Model: nil}

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
				user = tok.GetUserByToken(appContext.Db)
				if expired_to.Add(time.Minute).Before(time.Now()) {
					memcache.Set(token, user, time.Minute)
				}

			}

			if user == nil {
				response := UserJSON{Err: "incorrect user_id or tokens lifetime is expired",
					Model: nil}
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
				ctx := context.WithValue(req.Context(), user_key, user)
				next.ServeHTTP(w, req.WithContext(ctx))
			}

		} else {
			response := UserJSON{Err: "method not allowed",
				Model: nil}
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
		acc_l := req.Context().Value(accLoggerKey).(*log.Logger)
		sys_l := req.Context().Value(sysLoggerKey).(*log.Logger)
		if err != nil {
			sys_l.Printf("error during reading bot request: %s", err)
			logHttpRequest(acc_l, req, http.StatusOK)
			w.WriteHeader(http.StatusOK)
			return
		}

		var update TGBotApi.Update
		err = json.Unmarshal(body, &update)
		if err != nil {
			sys_l.Printf("error during unmarshaling request: %s: %s", req.URL.String(), err)
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
			sys_l.Printf("Max failed updates number exceeded on %d", upid)
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
		err = chat.CreateIfNotExists(appContext.Db)
		if err != nil {
			logHttpRequest(acc_l, req, http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(req.Context(), messageKey, message)

		next.ServeHTTP(w, req.WithContext(ctx))
		return
	})
}
