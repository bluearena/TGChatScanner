package requestHandler

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/clarifaiApi"
	"log"
	"net/http"
)

type RequestHandler struct {
	mux     *http.ServeMux
	Context *AppContext
}

type AppContext struct {
	Db            *gorm.DB
	Downloaders   *FileDownloadersPool
	PhotoHandlers *PhotoHandlersPool
	CfApi         *clarifaiApi.ClarifaiApi
	Cache         *MemoryCache
	Logger        *log.Logger
}

type appContext string

const appContextKey string = "appctx"

func NewRequestHandler() *RequestHandler {
	mux := http.NewServeMux()
	return &RequestHandler{mux: mux}
}

func (r *RequestHandler) RegisterHandlers() {
	r.mux.Handle("/api/v1/users.register", middleware(http.HandlerFunc(registerUser), *r.Context))
	r.mux.Handle("/api/v1/users.login", middleware(http.HandlerFunc(loginUser), *r.Context))
	r.mux.Handle("/api/v1/images.get", middleware(http.HandlerFunc(getImages), *r.Context))
	r.mux.Handle("/api/v1/images.restore", middleware(http.HandlerFunc(restoreImages), *r.Context))
	r.mux.Handle("/api/v1/images.remove", middleware(http.HandlerFunc(removeImages), *r.Context))
	r.mux.Handle("/api/v1/subs.remove", middleware(http.HandlerFunc(removeSubs), *r.Context))
	r.mux.Handle("/api/v1/chats.get", middleware(http.HandlerFunc(getChats), *r.Context))
	r.mux.Handle("/api/v1/tags.get", middleware(http.HandlerFunc(getTags), *r.Context))
	r.mux.Handle(TGBotApi.GetWebhookUrl(), middleware(http.HandlerFunc(BotUpdateHanlder), *r.Context))
}

func (r *RequestHandler) SetAppContext(context *AppContext) {
	r.Context = context
}

func AddAppContext(appctx AppContext, ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, appContextKey, appctx)
}

func middleware(next http.Handler, context AppContext) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := AddAppContext(context, req.Context(), req)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func (r *RequestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
	return
}
