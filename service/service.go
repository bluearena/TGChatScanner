package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"sync"
	"github.com/zwirec/TGChatScanner/requestHandler"
	"os/signal"
	"syscall"
	"log"
)

type Config map[string]map[string]interface{}

//Service s
type Service struct {
	mux         *http.ServeMux
	srv         *http.Server
	rAPIHandler *requestHandler.RequestHandler
	config      Config
}

func NewService() *Service {
	return &Service{rAPIHandler: requestHandler.NewRequestHandler(), mux:http.NewServeMux()}
}

func (s *Service) Run() error {

	usr, err := user.Current()
	if err != nil {
		return err
	}

	configPath := usr.HomeDir + "/.config/vkchatscanner/config.json"
	if err := s.parseConfig(configPath); err != nil {
		return err
	}
	s.srv = &http.Server{Addr: ":" + s.config["server"]["port"].(string), Handler:s.rAPIHandler}
	//s.signalProcessing()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		//defer wg.Done()
		if err := s.srv.ListenAndServe( /*":" + s.config["server"]["port"].(string), mux*/); err != nil {
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

func (s *Service) signalProcessing() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	go s.handler(c)
}

func (s Service) handler(c chan os.Signal) {
	for {
		<-c
		log.Print("Gracefully stopping...")
		s.srv.Close()
		os.Exit(0)
	}
}
