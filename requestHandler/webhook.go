package requestHandler

import (
	"encoding/json"
	"fmt"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"io/ioutil"
	"net/http"
	"regexp"
	"log"
)

func BotUpdateHanlder(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	logger := req.Context().Value(loggerContextKey).(*log.Logger)
	if err != nil {
		logger.Printf("Error during handling request on %s : %s", req.URL.String(), err)
		return
	}

	var update TGBotApi.Update
	err = json.Unmarshal(body, &update)
	toLog,_ := json.MarshalIndent(body,"","      ")
	logger.Printf("%+v", toLog)
	if err != nil {
		logger.Printf("Error during unmarshaling request on %s : %s", req.URL.String(), err)
		return
	}
	if pl := len(update.Message.Photo);pl != 0 {
		photo := update.Message.Photo[pl - 1]
		ctx := map[string]interface{}{}
		ctx["From"] = update.Message.From
		fb := &FileBasic{
			FileId:photo.FileId,
			Type:"photo",
			Context:ctx,
		}
		appContext.DownloadRequests <- fb
	} else if update.Message.Entities[0].Type == "bot_command" {
		if err := AddSubsription(update.Message, appContext.Cache); err != nil {
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
