package requestHandler

import (
	"encoding/json"
	"fmt"
	"github.com/zwirec/TGChatScanner/models"
	"net/http"
	"net/url"
	"strconv"
)

func middlewareLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		l := appContext.Logger

		if req.Method == "GET" {
			values, err := url.ParseQuery(req.URL.RawQuery)
			if err != nil {

				response := ResponseJSON{Err: "incorrect query rows",
					Model: nil}

				responseJSON, err := json.Marshal(response)
				if err != nil {
					writeResponse(w, string(responseJSON), http.StatusBadRequest)
					l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusBadRequest)
					return
				} else {
					l.Println(err)
					return
				}
			}
			result, ok := validateGETParams(values)
			if !ok {
				response := ResponseJSON{Err: "incorrect params",
					Model: nil}
				responseJSON, err := json.Marshal(response)
				if err != nil {
					writeResponse(w, string(responseJSON), http.StatusBadRequest)
					l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusBadRequest)
					return
				} else {
					l.Println(err)
					return
				}
			}

			user_chat := models.User_Chat{UserID: result["user_id"].(uint64),
				ChatID: result["chat_id"].(uint64),
				Token:  result["token"].(string),
			}

			if !user_chat.Validate(appContext.Db) {
				response := ResponseJSON{Err: "incorrect (user_id, chat_id) or tokens lifetime is expired",
					Model: nil}
				responseJSON, err := json.Marshal(response)
				if err != nil {
					writeResponse(w, string(responseJSON), http.StatusTeapot)
					l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusTeapot)
					return
				} else {
					l.Println(err)
					return
				}
			} else {
				next.ServeHTTP(w, req)
			}

		} else {
			response := ResponseJSON{Err: "method not allowed",
				Model: nil}
			responseJSON, err := json.Marshal(response)
			if err != nil {
				writeResponse(w, string(responseJSON), http.StatusMethodNotAllowed)
				return
			} else {
				l.Println(err)
			}
		}

	})
}

func validateGETParams(values url.Values) (map[string]interface{}, bool) {
	if values["user_id"] == nil || values["chat_id"] == nil || values["token"] == nil {
		return nil, false
	}
	result := map[string]interface{}{}

	var err error
	result["user_id"], err = strconv.ParseUint(values["user_id"][0], 10, 64)
	result["chat_id"], err = strconv.ParseUint(values["client_id"][0], 10, 64)

	if err != nil {
		return nil, false
	}
	result["token"] = values["token"]
	return result, false
}

func validateLoginParams(values map[string]interface{}) (ok bool) {
	if values["username"] == nil || values["password"] == nil {
		return false
	}
	return true
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
