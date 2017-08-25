package service

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
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

const (
	DefaultWorkersNumber   = 5
	DefaultCacheExpiraiton = 5 * time.Minute
	DefaultCacheClean      = 10 * time.Minute
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

type Service struct {
	sock         net.Listener
	mux          *http.ServeMux
	srv          *http.Server
	reqHandler   *requestHandler.RequestHandler
	config       Config
	errLogger    *log.Logger
	accessLogger *log.Logger
	notifier     chan os.Signal
	poolsWG      sync.WaitGroup
	endpointWg   sync.WaitGroup
	poolsDone    chan struct{}
}

func NewService() *Service {
	accL, errL := createLoggers()
	return &Service{
		reqHandler:   requestHandler.NewRequestHandler(),
		mux:          http.NewServeMux(),
		notifier:     make(chan os.Signal),
		poolsDone:    make(chan struct{}),
		errLogger:    errL,
		accessLogger: accL,
	}
}

func (s *Service) Run() error {

	if err := s.parseConfig(configURL); err != nil {
		s.errLogger.Println(err)
		return err
	}

	s.signalProcessing()

	db, err := s.initModels()

	if err != nil {
		s.errLogger.Println(err)
		return err
	}

	hostname := s.getHostname()

	clAPI := s.initClarifaiAPI()

	botAPI := s.initBotApi()

	workersNumber := s.getWorkersNumber()

	cache := memcache.New(DefaultCacheExpiraiton, DefaultCacheClean)

	imgPath, err := s.createImgPath()
	if err != nil {
		s.errLogger.Println(err)
		return err
	}

	downloadRequests := s.initPools(workersNumber)

	context := appContext.AppContext{
		DB:               db,
		DownloadRequests: downloadRequests,
		BotAPI:           botAPI,
		CfAPI:            clAPI,
		Cache:            cache,
		ErrLogger:        s.errLogger,
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
		s.errLogger.Println(err)
		os.Remove(s.config["server"]["socket"].(string))
		s.sock, _ = net.Listen("unix", s.config["server"]["socket"].(string))
	}
	if err := os.Chmod(s.config["server"]["socket"].(string), 0777); err != nil {
		s.errLogger.Println(err)
		s.notifier <- syscall.SIGINT
	}

	s.errLogger.Println("Socket opened")
	s.errLogger.Println("Server started")
	log.Println("Server started")
	if err := s.srv.Serve(s.sock); err != nil {
		s.errLogger.Println(err)
	}
	return nil
}

func (s *Service) parseConfig(URL string) error {
	var configRaw []byte

	_, err := url.Parse(URL)

	if err == nil {
		res, err := http.Get(URL)
		if err != nil {
			s.errLogger.Println(err)
			return err
		}

		configRaw, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			s.errLogger.Println(err)
			return err
		}

	} else {
		var err error
		configRaw, err = ioutil.ReadFile(URL)
		if err != nil {
			s.errLogger.Println(err)
			return err
		}
	}

	if err := json.Unmarshal(configRaw, &s.config); err != nil {
		s.errLogger.Println(err)
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
		s.errLogger.Println("Gracefully stopping...")
		log.Println("Gracefully stopping...")
		close(appContext.DownloadRequests)
		close(s.poolsDone)
		s.poolsWG.Wait()
		if err := s.srv.Shutdown(nil); err != nil {
			s.errLogger.Println(err)
			return
		}
	}
}

func (s *Service) initPools(workersNumber int) chan *file.FileBasic {
	dr := make(chan *file.FileBasic, workersNumber*2)

	fp := &requestHandler.FilePreparationsPool{In: dr, Done: s.poolsDone, WorkersNumber: workersNumber}
	fpOut := fp.Run(workersNumber*2, &s.poolsWG)

	forker := &forkers.ForkersPool{
		In:             fpOut,
		Done:           s.poolsDone,
		WorkersNumber:  workersNumber,
		ForkToFileInfo: requestHandler.CastToFileInfo,
		ForkToFileLink: requestHandler.CastToFileLink,
	}

	fdIn, prIn := forker.Run(workersNumber, workersNumber, &s.poolsWG)

	fd := &requestHandler.FileDownloadersPool{In: fdIn, Done: s.poolsDone, WorkersNumber: workersNumber}
	fdOut := fd.Run(workersNumber, &s.poolsWG)

	pr := &recognizers.PhotoRecognizersPool{In: prIn, Done: s.poolsDone, WorkersNumber: workersNumber}
	prOut := pr.Run(workersNumber, &s.poolsWG)

	deforker := &deforkers.DeforkersPool{
		In1:              fdOut,
		In2:              prOut,
		WorkersNumber:    workersNumber,
		DeforkDownloaded: requestHandler.CastFromDownloadedFile,
		DeforkRecognized: requestHandler.CastFromRecognizedPhoto,
	}

	dbsIn := deforker.Run(workersNumber*2, &s.poolsWG)

	dbs := &requestHandler.DbStoragesPool{In: dbsIn, WorkersNumber: workersNumber}
	dbs.Run(&s.poolsWG)
	return dr
}

func (s *Service) initClarifaiAPI() *clarifaiAPI.ClarifaiAPI {
	key := s.config["clarifai"]["api_key"].(string)
	url := s.config["clarifai"]["url"].(string)
	clAPI := clarifaiAPI.NewClarifaiAPI(key, url)
	return clAPI
}

func (s *Service) initBotApi() *TGBotAPI.BotAPI {
	key := s.config["tg_bot_api"]["token"].(string)
	return TGBotAPI.NewBotAPI(key)
}

func (s *Service) getWorkersNumber() int {

	wn, ok := s.config["server"]["workers"].(int)

	if !ok {
		wn = DefaultWorkersNumber
	}
	return wn
}

func createLoggers() (accLog *log.Logger, errLog *log.Logger) {
	errorlog, err := os.OpenFile("error.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		errorlog = os.Stderr
	}
	accesslog, err := os.OpenFile("access.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		accesslog = os.Stdout
	}
	accLog = log.New(errorlog, "", log.LstdFlags|log.Llongfile)
	errLog = log.New(accesslog, "", log.LstdFlags)
	return accLog, errLog
}

func (s *Service) initModels() (*gorm.DB, error) {
	db, err := modelManager.ConnectToDB(s.config["db"])
	if err != nil {
		return nil, err
	}

	if err := modelManager.InitDB(db); err != nil {
		return nil, err
	}
	return db, err
}

func (s *Service) createImgPath() (string, error) {
	imgPath, ok := s.config["chatscanner"]["images_path"].(string)

	if !ok {
		wd, _ := os.Getwd()
		imgPath = wd + "/uploads/"
	}

	if err := os.MkdirAll(imgPath, os.ModePerm); err != nil {
		return "", err
	}
	return imgPath, nil
}

func (s *Service) getHostname() string {
	hostname, _ := s.config["chatscanner"]["host"].(string)

	_, err := url.Parse(hostname)
	if err != nil {
		hostname, err = os.Hostname()
		if err != nil {
			s.errLogger.Println(err)
			hostname = "localhost"
		}
	}
	return hostname
}
