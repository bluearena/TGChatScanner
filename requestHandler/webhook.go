package requestHandler

import (
    "net/http"
    "io/ioutil"
    "log"
    "encoding/json"
    "github.com/zwirec/TGChatScanner/TGBotApi"
    "regexp"
)

func BotUpdateHanlder(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Response.Body)
    if err != nil {
        log.Printf("Error during handling request on %s : %s", r.URL.String(), err)
        return
    }

    var update TGBotApi.Update
    err = json.Unmarshal(body, &update)
    if err != nil {
        log.Printf("Error during unmarshaling request on %s : %s", r.URL.String(), err)
        return
    }
    ctx := r.Context()
    if len(update.Message.Photo) != 0 {

        for _, p := range update.Message.Photo {
            appCtx := ctx.Value(appContextKey).(AppContext)
            appCtx.PhotoHandlers.RequestPhotoHandling(p)
        }
    } else if update.Message.Entities[0].Type == "bot_command" {
        appCtx := ctx.Value(appContextKey).(AppContext)
        AddSubsription(update.Message, appCtx.Cache)
    }
    w.WriteHeader(http.StatusOK)
}

func AddSubsription(message TGBotApi.Message, cache *MemoryCache) {
    r := regexp.MustCompile(`\/(startgroup|start)?\s+(?P<token>[[:alnum:]]+)`)
    command := r.FindStringSubmatch(message.Text)

    if len(command) == 0 {
        log.Printf("unexpected command %s", message.Text)
        return
    }

    userKey := command[2]
    userId, ok := cache.Get(userKey)
    if !ok {
        log.Printf("user not found, key %s", userKey)
    }

    chatId := message.From.Id

    //TODO: store subscripton in database
    log.Printf("New subscription: %s, $s", userId, chatId)
}
