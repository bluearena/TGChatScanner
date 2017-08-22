package service

import (
	"encoding/json"
	"fmt"
	memcache "github.com/patrickmn/go-cache"
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
	"time"
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
	sock         net.Listener
	mux          *http.ServeMux
	srv          *http.Server
	rAPIHandler  *requestHandler.RequestHandler
	config       Config
	sysLogger    *log.Logger
	accessLogger *log.Logger
	notifier     chan os.Signal
	poolsWG      sync.WaitGroup
	poolsDone    chan struct{}
}

func NewService() *Service {
	return &Service{
		rAPIHandler: requestHandler.NewRequestHandler(),
		mux:         http.NewServeMux(),
		notifier:    make(chan os.Signal),
		poolsDone:   make(chan struct{}),
	}
}

func (s *Service) Run() error {

	errorlog, err := os.OpenFile("error.log", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		errorlog = os.Stderr
	}
	accesslog, err := os.OpenFile("access.log", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		accesslog = os.Stdout
	}

	s.sysLogger = log.New(errorlog, "", log.LstdFlags|log.Llongfile)
	s.accessLogger = log.New(accesslog, "", log.LstdFlags)

	if err := s.parseConfig(configUrl); err != nil {
		s.sysLogger.Println(err)
		return err
	}

	s.signalProcessing()

	db, err := modelManager.ConnectToDB(s.config["db"])

	if err != nil {
		s.sysLogger.Println(err)
		return err
	}

	if err := modelManager.InitDB(db); err != nil {
		s.sysLogger.Println(err)
	}

	clApi := clarifaiApi.NewClarifaiApi(s.config["clarifai"]["api_key"].(string))

	botApi := TGBotApi.NewBotApi(s.config["tg_bot_api"]["token"].(string))

	workers_n, ok := s.config["server"]["workers"].(int)

	if !ok {
		workers_n = 10
	}

	dr := make(chan *requestHandler.FileBasic, workers_n*2)

	poolStopper := make(chan struct{})

	fp := &requestHandler.FilePreparatorsPool{In: dr, Done: poolStopper, WorkersNumber: workers_n}
	fpOut := fp.Run(workers_n*2, s.poolsWG)

	forker := &requestHandler.ForkersPool{
		In:             fpOut,
		Done:           poolStopper,
		WorkersNumber:  workers_n,
		ForkToFileInfo: requestHandler.CastToFileInfo,
		ForkToFileLink: requestHandler.CastToFileLink,
	}

	fdIn, prIn := forker.Run(workers_n, workers_n, s.poolsWG)

	fd := &requestHandler.FileDownloadersPool{In: fdIn, Done: poolStopper, WorkersNumber: workers_n}
	fdOut := fd.Run(workers_n, s.poolsWG)

	pr := &requestHandler.PhotoRecognizersPool{In: prIn, Done: poolStopper, WorkersNumber: workers_n}
	prOut := pr.Run(workers_n, s.poolsWG)

	deforker := &requestHandler.DeforkersPool{
		In1:              fdOut,
		In2:              prOut,
		WorkersNumber:    workers_n,
		DeforkDownloaded: requestHandler.CastFromDownloadedFile,
		DeforkRecognized: requestHandler.CastFromRecognizedPhoto,
	}

	dbsIn := deforker.Run(workers_n*2, s.poolsWG)

	dbs := &requestHandler.DbStoragersPool{In: dbsIn, WorkersNumber: workers_n}
	dbs.Run(s.poolsWG)

	cache := memcache.New(5*time.Minute, 10*time.Minute)

	imgPath, ok := s.config["chatscanner"]["images_path"].(string)

	if err := os.MkdirAll(imgPath, os.ModePerm); err != nil {
		s.sysLogger.Println(err)
	}

	hostname, ok := s.config["chatscanner"]["host"].(string)

	_, err = url.Parse(hostname)

	if err != nil {
		hostname, err = os.Hostname()
		if err != nil {
			s.sysLogger.Println(err)
			hostname = "localhost"
		}
	}

	context := requestHandler.AppContext{
		Db:               db,
		DownloadRequests: dr,
		BotApi:           botApi,
		CfApi:            clApi,
		Cache:            cache,
		SysLogger:        s.sysLogger,
		AccessLogger:     s.accessLogger,
		ImagesPath:       imgPath,
		Hostname:         hostname,
	}

	s.rAPIHandler.SetAppContext(&context)
	s.rAPIHandler.RegisterHandlers()

	s.srv = &http.Server{Handler: s.rAPIHandler}

	defer close(poolStopper)

	var wg sync.WaitGroup
	wg.Add(1)

	go s.endpoint()

	wg.Wait()
	return nil
}

func (s *Service) endpoint() (err error) {

	s.sock, err = net.Listen("unix", s.config["server"]["socket"].(string))
	if err != nil {
		s.sysLogger.Println(err)
		s.notifier <- syscall.SIGINT
	}
	if err := os.Chmod(s.config["server"]["socket"].(string), 0777); err != nil {
		s.sysLogger.Println(err)
		s.notifier <- syscall.SIGINT
	}

	s.sysLogger.Println("Socket opened")
	s.sysLogger.Println("Server started")

	if err := s.srv.Serve(s.sock); err != nil {
		s.sysLogger.Println(err)
		//s.notifier <- syscall.SIGINT
	}
	return nil
}

func (s *Service) parseConfig(_url string) error {
	var configRaw []byte

	_, err := url.Parse(_url)

	if err == nil {
		res, err := http.Get(_url)
		if err != nil {
			s.sysLogger.Println(err)
			return err
		}

		configRaw, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			s.sysLogger.Println(err)
			return err
		}

	} else {
		var err error
		configRaw, err = ioutil.ReadFile(_url)
		if err != nil {
			s.sysLogger.Println(err)
			return err
		}
	}

	if err := json.Unmarshal(configRaw, &s.config); err != nil {
		s.sysLogger.Println(err)
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
		s.sysLogger.Println("Gracefully stopping...")
		close(s.poolsDone)
		s.poolsWG.Wait()
		if err := s.srv.Shutdown(nil); err != nil {
			s.sysLogger.Println(err)
			return
		}
		//if err := s.sock.Close(); err != nil {
		//	s.sysLogger.Println(err)
		//}
		//if err := os.Remove(s.config["server"]["socket"].(string)); err != nil {
		//	s.sysLogger.Println(err)
		//	return
		//}
		os.Exit(0)
	}
}
