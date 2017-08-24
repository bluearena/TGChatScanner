package service

import (
	"encoding/json"
	"fmt"
	memcache "github.com/patrickmn/go-cache"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/clarifaiApi"
	"github.com/zwirec/TGChatScanner/modelManager"
	"github.com/zwirec/TGChatScanner/requestHandler"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	"github.com/zwirec/TGChatScanner/requestHandler/deforkers"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"github.com/zwirec/TGChatScanner/requestHandler/forkers"
	"github.com/zwirec/TGChatScanner/requestHandler/recognizers"
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
	configURL = os.Getenv("TGCHATSCANNER_REMOTE_CONFIG")
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
	if configURL == "" {
		configURL = home + "/.config/tgchatscanner/config.json"
	}
}

type Config map[string]map[string]interface{}

//Service s
type Service struct {
	sock         net.Listener
	mux          *http.ServeMux
	srv          *http.Server
	reqHandler   *requestHandler.RequestHandler
	config       Config
	ErrLogger    *log.Logger
	accessLogger *log.Logger
	notifier     chan os.Signal
	poolsWG      sync.WaitGroup
	endpointWg   sync.WaitGroup
	poolsDone    chan struct{}
}

func NewService() *Service {
	return &Service{
		reqHandler: requestHandler.NewRequestHandler(),
		mux:        http.NewServeMux(),
		notifier:   make(chan os.Signal),
		poolsDone:  make(chan struct{}),
	}
}

func (s *Service) Run() error {

	errorlog, err := os.OpenFile("error.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		errorlog = os.Stderr
	}
	accesslog, err := os.OpenFile("access.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		accesslog = os.Stdout
	}

	s.ErrLogger = log.New(errorlog, "", log.LstdFlags|log.Llongfile)
	s.accessLogger = log.New(accesslog, "", log.LstdFlags)

	if err := s.parseConfig(configURL); err != nil {
		s.ErrLogger.Println(err)
		return err
	}

	s.signalProcessing()

	db, err := modelManager.ConnectToDB(s.config["db"])

	if err != nil {
		s.ErrLogger.Println(err)
		return err
	}

	if err := modelManager.InitDB(db); err != nil {
		s.ErrLogger.Println(err)
	}

	clAPI := clarifaiAPI.NewClarifaiAPI(s.config["clarifai"]["api_key"].(string))

	botAPI := TGBotAPI.NewBotAPI(s.config["tg_bot_api"]["token"].(string))

	workers_n, ok := s.config["server"]["workers"].(int)

	if !ok {
		workers_n = 10
	}

	dr := make(chan *file.FileBasic, workers_n*2)

	fp := &requestHandler.FilePreparationsPool{In: dr, Done: s.poolsDone, WorkersNumber: workers_n}
	fpOut := fp.Run(workers_n*2, &s.poolsWG)

	forker := &forkers.ForkersPool{
		In:             fpOut,
		Done:           s.poolsDone,
		WorkersNumber:  workers_n,
		ForkToFileInfo: requestHandler.CastToFileInfo,
		ForkToFileLink: requestHandler.CastToFileLink,
	}

	fdIn, prIn := forker.Run(workers_n, workers_n, &s.poolsWG)

	fd := &requestHandler.FileDownloadersPool{In: fdIn, Done: s.poolsDone, WorkersNumber: workers_n}
	fdOut := fd.Run(workers_n, &s.poolsWG)

	pr := &recognizers.PhotoRecognizersPool{In: prIn, Done: s.poolsDone, WorkersNumber: workers_n}
	prOut := pr.Run(workers_n, &s.poolsWG)

	deforker := &deforkers.DeforkersPool{
		In1:              fdOut,
		In2:              prOut,
		WorkersNumber:    workers_n,
		DeforkDownloaded: requestHandler.CastFromDownloadedFile,
		DeforkRecognized: requestHandler.CastFromRecognizedPhoto,
	}

	dbsIn := deforker.Run(workers_n*2, &s.poolsWG)

	dbs := &requestHandler.DbStoragesPool{In: dbsIn, WorkersNumber: workers_n}
	dbs.Run(&s.poolsWG)

	cache := memcache.New(5*time.Minute, 10*time.Minute)

	imgPath, ok := s.config["chatscanner"]["images_path"].(string)

	if !ok {
		wd, _ := os.Getwd()
		imgPath = wd + "/uploads/"
	}

	if err := os.MkdirAll(imgPath, os.ModePerm); err != nil {
		s.ErrLogger.Println(err)
	}

	hostname, ok := s.config["chatscanner"]["host"].(string)

	_, err = url.Parse(hostname)

	if err != nil {
		hostname, err = os.Hostname()
		if err != nil {
			s.ErrLogger.Println(err)
			hostname = "localhost"
		}
	}

	context := appContext.AppContext{
		DB:               db,
		DownloadRequests: dr,
		BotAPI:           botAPI,
		CfAPI:            clAPI,
		Cache:            cache,
		ErrLogger:        s.ErrLogger,
		AccessLogger:     s.accessLogger,
		ImagesPath:       imgPath,
		Hostname:         hostname,
	}

	appContext.SetAppContext(&context)
	s.reqHandler.RegisterHandlers()

	s.srv = &http.Server{Handler: s.reqHandler}

	s.endpointWg.Add(1)

	go func() {
		defer s.endpointWg.Done()
		s.endpoint()

	}()
	s.endpointWg.Wait()
	return nil
}

func (s *Service) endpoint() (err error) {
	s.sock, err = net.Listen("unix", s.config["server"]["socket"].(string))
	if err != nil {
		s.ErrLogger.Println(err)
		os.Remove(s.config["server"]["socket"].(string))
		s.sock, _ = net.Listen("unix", s.config["server"]["socket"].(string))
	}
	if err := os.Chmod(s.config["server"]["socket"].(string), 0777); err != nil {
		s.ErrLogger.Println(err)
		s.notifier <- syscall.SIGINT
	}

	s.ErrLogger.Println("Socket opened")
	s.ErrLogger.Println("Server started")
	log.Println("Server started")
	if err := s.srv.Serve(s.sock); err != nil {
		s.ErrLogger.Println(err)
	}
	return nil
}

func (s *Service) parseConfig(URL string) error {
	var configRaw []byte

	_, err := url.Parse(URL)

	if err == nil {
		res, err := http.Get(URL)
		if err != nil {
			s.ErrLogger.Println(err)
			return err
		}

		configRaw, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			s.ErrLogger.Println(err)
			return err
		}

	} else {
		var err error
		configRaw, err = ioutil.ReadFile(URL)
		if err != nil {
			s.ErrLogger.Println(err)
			return err
		}
	}

	if err := json.Unmarshal(configRaw, &s.config); err != nil {
		s.ErrLogger.Println(err)
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
		s.ErrLogger.Println("Gracefully stopping...")
		log.Println("Gracefully stopping...")
		close(appContext.DownloadRequests)
		close(s.poolsDone)
		s.poolsWG.Wait()
		if err := s.srv.Shutdown(nil); err != nil {
			s.ErrLogger.Println(err)
			return
		}
	}
}
