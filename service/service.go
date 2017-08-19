package service

import (
	"encoding/json"
	"github.com/zwirec/TGChatScanner/clarifaiApi"
	"github.com/zwirec/TGChatScanner/modelManager"
	"github.com/zwirec/TGChatScanner/requestHandler"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"sync"
	"syscall"
	"github.com/zwirec/TGChatScanner/TGBotApi"
)

type Config map[string]map[string]interface{}

//Service s
type Service struct {
	mux         *http.ServeMux
	srv         *http.Server
	rAPIHandler *requestHandler.RequestHandler
	config      Config
	logger      *log.Logger
}

func NewService() *Service {
	return &Service{
		rAPIHandler: requestHandler.NewRequestHandler(),
		mux:         http.NewServeMux(),
		logger:      log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (s *Service) Run() error {
	configUrl := os.Getenv("TGCHATSCANNER_REMOTE_CONFIG")
	rc := true

	if configUrl == "" {
		s.logger.Println("Using local config")

		rc = false
		usr, err := user.Current()

		if err != nil {
			s.logger.Println(err)
			return err
		}

		configUrl = usr.HomeDir + "/.config/tgchatscanner/config.json"
	} else {
		s.logger.Println("Using remote config")
	}

	if err := s.parseConfig(configUrl, rc); err != nil {
		s.logger.Println(err)
		return err
	}

	s.signalProcessing()

	db, err := modelManager.ConnectToDB(s.config["db"])
	if err != nil {
		s.logger.Println(err)
		return err
	}

	clApi := clarifaiApi.NewClarifaiApi(s.config["clarifai"]["api_key"].(string))

	botApi := TGBotApi.NewBotApi(s.config["tg_bot_api"]["token"].(string))

	workers_n, ok := s.config["server"]["workers"].(int)

	if !ok {
		workers_n = 10
	}
    fdp := requestHandler.NewFileDownloaderPool(workers_n, 100)

    php := requestHandler.NewPhotoHandlersPool(workers_n, 100)

	cache := requestHandler.MemoryCache{}
	context := requestHandler.AppContext{
		Db:            db,
		Downloaders:   fdp,
		PhotoHandlers: php,
		BotApi:        botApi,
		CfApi:         clApi,
		Cache:         &cache,
		Logger:        s.logger,
	}

	s.rAPIHandler.SetAppContext(&context)
	s.rAPIHandler.RegisterHandlers()

	s.srv = &http.Server{Handler: s.rAPIHandler}

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		l, err := net.Listen("unix", s.config["server"]["socket"].(string))
		if err != nil {
			s.logger.Println(err)
			wg.Done()
		}
		s.logger.Println("Socket opened")
		defer os.Remove(s.config["server"]["socket"].(string))
		defer l.Close()

		log.Println("Server started")
		if err := s.srv.Serve(l); err != nil {
			s.logger.Println(err)
			wg.Done()
		}
	}()

	wg.Wait()
	return nil
}

func (s *Service) parseConfig(url string, remote bool) error {
	var configRaw []byte

	if remote {
		res, err := http.Get(url)
		if err != nil {
			s.logger.Println(err)
			return err
		}

		configRaw, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			s.logger.Println(err)
			return err
		}

	} else {
		var err error

		configRaw, err = ioutil.ReadFile(url)
		if err != nil {
			s.logger.Println(err)
			return err
		}
	}

	if err := json.Unmarshal(configRaw, &s.config); err != nil {
		s.logger.Println(err)
		return err
	}

	return nil

}

func (s *Service) signalProcessing() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	go s.handler(c)
}

func (s *Service) handler(c chan os.Signal) {
	for {
		<-c
		log.Print("Gracefully stopping...")
		s.srv.Shutdown(nil)
		os.Exit(0)
	}
}
