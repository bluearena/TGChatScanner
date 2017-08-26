package requestHandler

import (
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	. "github.com/zwirec/TGChatScanner/requestHandler/handlers"
	"net/http"
)

type RequestHandler struct {
	mux *http.ServeMux
}

func NewRequestHandler() *RequestHandler {
	mux := http.NewServeMux()
	return &RequestHandler{mux: mux}
}

func (r *RequestHandler) RegisterHandlers() {
	r.mux.Handle("/api/v1/images.get", MiddlewareLogin(http.HandlerFunc(GetImages)))
	r.mux.Handle("/api/v1/chats.get", MiddlewareLogin(http.HandlerFunc(GetChats)))
	r.mux.Handle("/api/v1/chat.tags", MiddlewareLogin(http.HandlerFunc(GetChatTags)))
	r.mux.Handle("/api/v1/users.tags", MiddlewareLogin(http.HandlerFunc(GetUserTags)))
	r.mux.Handle("/"+appContext.BotAPI.Token, ExtractUpdate(
		BotRouter(
			http.HandlerFunc(BotUpdateHandler))))
}

func (r *RequestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
