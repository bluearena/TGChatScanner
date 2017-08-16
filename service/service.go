package service

import (
	"encoding/json"
	"fmt"
	"github.com/zwirec/TGChatScanner/modelManager"
	"net/http"
	"os"
	"os/user"
	"sync"
)

type Config map[string]map[string]interface{}

const (
	filenameConfig = "/Users/zwirec/.config/vkchatscanner/config.json"
)

//Service s
type Service struct {
	srv    *http.Server
	config Config
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Run() error {

	if err := s.parseConfig(filenameConfig); err != nil {
		return err
	}

	s.srv = &http.Server{Addr: ":" + (s.config["server"]["port"]).(string)}

	if err := modelManager.ConnectToDB(s.config["db"]); err != nil {
		return err
	}

	/* Вынести в отдельный файл хендлеры*/

	//s.registerHandleFuncs()

	//s.signalProcessing()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		//defer wg.Done()
		if err := s.srv.ListenAndServe(); err != nil {
			wg.Done()
		}
	}()

	wg.Wait()
	return nil
}

func (s *Service) parseConfig(filename string) error {
	file, err := os.Open(filename)

	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)

	decoder.Decode(&s.config)

	if err != nil {
		return fmt.Errorf("%q: incorrect configuration file", filename)
	}
	return nil
}
