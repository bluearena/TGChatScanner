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
	"crypto/tls"
	"github.com/kabukky/httpscerts"
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
	return &Service{rAPIHandler: requestHandler.NewRequestHandler(), mux: http.NewServeMux()}
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

	s.signalProcessing()

	err = httpscerts.Check("cert.pem", "key.pem")
	// If they are not available, generate new ones.
	if err != nil {
		err = httpscerts.Generate("cert.pem", "key.pem", "127.0.0.1:"+ s.config["server"]["port"].(string))
		if err != nil {
			log.Fatal("Error: Couldn't create https certs.")
		}
	}

	cer, err := tls.LoadX509KeyPair("cert.pem", "key.pem")

	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	s.srv = &http.Server{Addr: ":" + s.config["server"]["port"].(string), Handler: s.rAPIHandler, TLSConfig:config}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		//defer wg.Done()
		if err := s.srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
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

func (s *Service) handler(c chan os.Signal) {
	for {
		<-c
		log.Print("Gracefully stopping...")
		s.srv.Shutdown(nil)
		os.Exit(0)
	}
}
