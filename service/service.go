package service

import (
	"encoding/json"
	"fmt"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/clarifaiApi"
	"github.com/zwirec/TGChatScanner/modelManager"
	"github.com/zwirec/TGChatScanner/requestHandler"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"os/user"
	"sync"
	"syscall"
)

var (
	home      = os.Getenv("HOME")
	configUrl = os.Getenv("TGCHATSCANNER_REMOTE_CONFIG")
)

func init() {
	if home == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		home := u.HomeDir
		fmt.Fprint(ioutil.Discard, home)
	}
	if configUrl == "" {
		configUrl = home + "/.config/tgchatscanner/config.json"
	}
}

type Config map[string]map[string]interface{}

//Service s
type Service struct {
	sock        net.Listener
	mux         *http.ServeMux
	srv         *http.Server
	rAPIHandler *requestHandler.RequestHandler
	config      Config
	logger      *log.Logger
	notifier    chan os.Signal
}

func NewService() *Service {
	return &Service{
		rAPIHandler: requestHandler.NewRequestHandler(),
		mux:         http.NewServeMux(),
		logger:      log.New(os.Stdout, "", log.LstdFlags),
		notifier:    make(chan os.Signal),
	}
}

func (s *Service) Run() error {

	if err := s.parseConfig(configUrl); err != nil {
		s.logger.Println(err)
		return err
	}

	s.signalProcessing()

	db, err := modelManager.ConnectToDB(s.config["db"])

	if err != nil {
		s.logger.Println(err)
		return err
	}

	modelManager.InitDB(db)

	clApi := clarifaiApi.NewClarifaiApi(s.config["clarifai"]["api_key"].(string))

	botApi := TGBotApi.NewBotApi(s.config["tg_bot_api"]["token"].(string))

	workers_n, ok := s.config["server"]["workers"].(int)

	if !ok {
		workers_n = 10
	}

	dr := make(chan *requestHandler.FileBasic, workers_n*2)

	poolStoper := make(chan struct{})

	fp := &requestHandler.FilePreparatorsPool{In: dr, Done: poolStoper, WorkersNumber: workers_n}
	fpOut := fp.Run(workers_n * 2)

	forker := &requestHandler.ForkersPool{
		In:             fpOut,
		Done:           poolStoper,
		WorkersNumber:  workers_n,
		ForkToFileInfo: requestHandler.CastToFileInfo,
		ForkToFileLink: requestHandler.CastToFileLink,
	}

	fdIn, prIn := forker.Run(workers_n, workers_n)

	fd := &requestHandler.FileDownloadersPool{In: fdIn, Done: poolStoper, WorkersNumber: workers_n}
	fdOut := fd.Run(workers_n)

	pr := &requestHandler.PhotoRecognizersPool{In: prIn, Done: poolStoper, WorkersNumber: workers_n}
	prOut := pr.Run(workers_n)

	deforker := &requestHandler.DeforkersPool{
		In1:              fdOut,
		In2:              prOut,
		WorkersNumber:    workers_n,
		DeforkDownloaded: requestHandler.CastFromDownloadedFile,
		DeforkRecognized: requestHandler.CastFromRecognizedPhoto,
	}

	dbsIn := deforker.Run(workers_n * 2)

	dbs := &requestHandler.DbStoragersPool{In: dbsIn}

	dbs.Run()

	cache := requestHandler.MemoryCache{}
	context := requestHandler.AppContext{
		Db:               db,
		DownloadRequests: dr,
		PoolStop:         poolStoper,
		BotApi:           botApi,
		CfApi:            clApi,
		Cache:            &cache,
		Logger:           s.logger,
	}

	s.rAPIHandler.SetAppContext(&context)
	s.rAPIHandler.RegisterHandlers()

	s.srv = &http.Server{Handler: s.rAPIHandler}
	defer close(poolStoper)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.endpoint()
	wg.Wait()
	return nil
}

func (s *Service) endpoint() (err error) {

	s.sock, err = net.Listen("unix", s.config["server"]["socket"].(string))

	if err != nil {
		s.logger.Println(err)
		s.notifier <- syscall.SIGINT
	}
	if err := os.Chmod(s.config["server"]["socket"].(string), 0777); err != nil {
		s.logger.Println(err)
		s.notifier <- syscall.SIGINT
	}
	s.logger.Println("Socket opened")
	s.logger.Println("Server started")
	if err := s.srv.Serve(s.sock); err != nil {
		s.logger.Println(err)
		s.notifier <- syscall.SIGINT
	}
	return nil
}

func (s *Service) parseConfig(_url string) error {
	var configRaw []byte

	_, err := url.Parse(_url)

	if err != nil {
		res, err := http.Get(_url)
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
		configRaw, err = ioutil.ReadFile(_url)
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
	signal.Notify(s.notifier, syscall.SIGINT)
	go s.handler(s.notifier)
}

func (s *Service) handler(c chan os.Signal) {
	for {
		<-c
		log.Print("Gracefully stopping...")
		s.srv.Shutdown(nil)
		s.sock.Close()
		os.Remove(s.config["server"]["socket"].(string))
		os.Exit(0)
	}
}
