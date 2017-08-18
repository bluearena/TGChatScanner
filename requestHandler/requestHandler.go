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
    Logger        *log.Logger
}

type appContext string

const appContextKey string = "appctx"

func NewRequestHandler() *RequestHandler {
    mux := http.NewServeMux()
    return &RequestHandler{mux: mux}
}

func (r *RequestHandler) RegisterHandlers() {
    r.mux.Handle("/api/users/register", middleware(http.HandlerFunc(RegisterUser), *r.Context))
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

func RegisterUser(w http.ResponseWriter, req *http.Request) {
    req.ParseForm()
    data := req.PostForm
    hash, err := bcrypt.GenerateFromPassword([]byte(data["password"][0]), bcrypt.DefaultCost)
    logger := req.Context().Value("appctx").(AppContext).Logger

    if err != nil {
        logger.Println(err)
    }
    fmt.Fprintf(w, "Hash to store: %s", string(hash))
    return
}

func (r *RequestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.mux.ServeHTTP(w, req)
    return
}
