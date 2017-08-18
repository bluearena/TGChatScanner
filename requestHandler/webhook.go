package requestHandler

import (
    "net/http"
    "io/ioutil"
    "log"
    "encoding/json"
    "github.com/zwirec/TGChatScanner/TGBotApi"
    "regexp"
	"fmt"
)

func BotUpdateHanlder(w http.ResponseWriter, req *http.Request) {
    body, err := ioutil.ReadAll(req.Body)
    logger := req.Context().Value("appctx").(AppContext).Logger
    if err != nil {
        logger.Printf("Error during handling request on %s : %s", req.URL.String(), err)
        return
    }

    var update TGBotApi.Update
    err = json.Unmarshal(body, &update)
    if err != nil {
        logger.Printf("Error during unmarshaling request on %s : %s", req.URL.String(), err)
        return
    }
    ctx := req.Context()
    if len(update.Message.Photo) != 0 {

        for _, p := range update.Message.Photo {
            appCtx := ctx.Value(appContextKey).(AppContext)
            appCtx.PhotoHandlers.RequestPhotoHandling(p)
        }
    } else if update.Message.Entities[0].Type == "bot_command" {
        appCtx := ctx.Value(appContextKey).(AppContext)
        if err := AddSubsription(update.Message, appCtx.Cache); err != nil {
			logger.Println(err)
			return
		}
    }
    w.WriteHeader(http.StatusOK)
}

func AddSubsription(message TGBotApi.Message, cache *MemoryCache) error {
    r := regexp.MustCompile(`\/(startgroup|start)?\s+(?P<token>[[:alnum:]]+)`)
    command := r.FindStringSubmatch(message.Text)

    if len(command) == 0 {
		return fmt.Errorf("unexpected command %s", message.Text)
    }

    userKey := command[2]
    userId, ok := cache.Get(userKey)
    if !ok {
        return fmt.Errorf("user not found, key %s", userKey)
    }

    chatId := message.From.Id

    //TODO: store subscripton in database
	return fmt.Errorf("New subscription: %s, $s", userId, chatId)

}
