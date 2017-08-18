package requestHandler

import (
    "net/http"
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "log"
    "github.com/jinzhu/gorm"
    "github.com/zwirec/TGChatScanner/clarifaiApi"
    "context"
    "github.com/zwirec/TGChatScanner/TGBotApi"
)

type RequestHandler struct {
    mux     *http.ServeMux
    Context *AppContext
}

type AppContext struct {
    Db            *gorm.DB
    Downloaders   *FileDownloaderPool
    PhotoHandlers *PhotoHandlersPool
    CfApi         *clarifaiApi.ClarifaiApi
    Cache         *MemoryCache
}

type key int

const appContextKey = 0

func NewRequestHandler() *RequestHandler {
    mux := http.NewServeMux()
    // mux.HandleFunc("/api/users/register", RegisterUser)
    return &RequestHandler{mux: mux}
}

func (r *RequestHandler) RegisterHandlers() {
    r.mux.Handle("/api/users/register", middleware(http.HandlerFunc(RegisterUser), *r.Context))
    r.mux.Handle(TGBotApi.GetWebhookUrl(), middleware(http.HandlerFunc(BotUpdateHanlder), *r.Context))
}

func (rh *RequestHandler) SetAppContext(context *AppContext) {
    rh.Context = context
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

func RegisterUser(w http.ResponseWriter, req *http.Request) {
    req.ParseForm()
    data := req.PostForm
    hash, err := bcrypt.GenerateFromPassword([]byte(data["password"][0]), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Hash to store:", string(hash))
    return
}

func (rH *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    rH.mux.ServeHTTP(w, r)
    return
}
