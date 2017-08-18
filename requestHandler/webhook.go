package requestHandler

import (
<<<<<<< HEAD
	"net/http"
	"io/ioutil"
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
=======
    "net/http"
    "io/ioutil"
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
        appCtx := ctx.Value(appContextKey).(AppContext)
        appCtx.PhotoHandlers.RequestPhotoHandling(&update.Message, appCtx.CfApi)

    } else if update.Message.Entities[0].Type == "bot_command" {
        appCtx := ctx.Value(appContextKey).(AppContext)
        if err := AddSubsription(update.Message, appCtx.Cache); err != nil {
            logger.Println(err)
            return
        }
    }
    w.WriteHeader(http.StatusOK)
>>>>>>> 695fb6bbd05ea59c0c22028f23308d6b69a14f19
}

func AddSubsription(message TGBotApi.Message, cache *MemoryCache) error {
	r := regexp.MustCompile(`\/(startgroup|start)?\s+(?P<token>[[:alnum:]]+)`)
	command := r.FindStringSubmatch(message.Text)

<<<<<<< HEAD
	if len(command) == 0 {
		return fmt.Errorf("unexpected command %s", message.Text)
	}
=======
    if len(command) == 0 {
        return fmt.Errorf("unexpected command %s", message.Text)
    }
>>>>>>> 695fb6bbd05ea59c0c22028f23308d6b69a14f19

	userKey := command[2]
	userId, ok := cache.Get(userKey)
	if !ok {
		return fmt.Errorf("user not found, key %s", userKey)
	}

	chatId := message.From.Id

<<<<<<< HEAD
	//TODO: store subscripton in database
	return fmt.Errorf("New subscription: %s, $s", userId, chatId)
=======
    //TODO: store subscripton in database
    return fmt.Errorf("New subscription: %s, $s", userId, chatId)
>>>>>>> 695fb6bbd05ea59c0c22028f23308d6b69a14f19

}
