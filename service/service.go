package service

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"sync"
	"../modelManager"
	"os"
	"fmt"
)

type DBInfo map[string]string
type SInfo map[string]string

const (
	filenameServerConfig = "/Users/zwirec/.config/vkchatscanner/config.json"
	filenameDBConfig     = `/Users/zwirec/.config/vkchatscanner/db_config.json`
)

//Service s
type Service struct {
	srv    *http.Server
	dbInfo DBInfo
	sInfo  SInfo
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Run() error {

	s.srv = &http.Server{Addr: ":" + s.sInfo["port"]}

	if err := s.parseServerConfig(filenameServerConfig); err != nil {
		return err
	}

	if err := s.parseDBConfig(filenameDBConfig); err != nil {
		return err
	}

	os.Setenv("PGHOST", s.dbInfo["host"])

	//....//

	if err := modelManager.ConnectToDB(); err != nil {
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

func (s *Service) parseServerConfig(filename string) error {

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(data), &s.dbInfo)

	if err != nil {
		return fmt.Errorf("%q: incorrect configuration file", filename)
	}
	return nil
}

func (s *Service) parseDBConfig(filename string) error {

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(data), &s.sInfo)

	if err != nil {
		return fmt.Errorf("%q: incorrect configuration file", filename)
	}
	return nil
}
