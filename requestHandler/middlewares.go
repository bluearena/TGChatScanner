package requestHandler

import (
	"context"
	"encoding/json"
	"github.com/zwirec/TGChatScanner/models"
	"net/http"
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

			tok := models.Token{Token: token}

			user := tok.GetUserByToken(appContext.Db)

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
