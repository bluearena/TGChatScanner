package requestHandler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zwirec/TGChatScanner/models"
	"net/http"
	"net/url"
	"strconv"
)

type UserJSON struct {
	Err   string       `json:"error,omitempty"`
	Model *models.User `json:"entity,omitempty"`
}

var user_key = "user"

func middlewareLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		l := appContext.Logger

		if req.Method == "GET" {
			token := req.Header.Get("X-User-Token")

			if token == "" {

				response := UserJSON{Err: "X-User-Token isn't ",
					Model: nil}

				responseJSON, err := json.Marshal(response)
				if err != nil {
					writeResponse(w, string(responseJSON), http.StatusForbidden)
					l.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, http.StatusForbidden)
					return
				} else {
					l.Println(err)
					return
				}
			}

			tok := models.Token{Token: token}

			user := tok.GetUser(appContext.Db)

			if user == nil {
				response := UserJSON{Err: "incorrect user_id or tokens lifetime is expired",
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
				ctx := context.WithValue(req.Context(), user_key, user)
				next.ServeHTTP(w, req.WithContext(ctx))
			}

		} else {
			response := UserJSON{Err: "method not allowed",
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

func getImages(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "images.get")
	return
}

func restoreImages(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "images.restore")
	return
}
func removeImages(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "images.remove")
	return
}
func getChats(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "chats.get")
	return
}
func getTags(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "tags.get")
	return
}
func removeSubs(w http.ResponseWriter, req *http.Request) {
	//TODO
	fmt.Fprint(w, "subs.remove")
	return
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
