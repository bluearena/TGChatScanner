package service

import (
	"encoding/json"
	"github.com/zwirec/TGChatScanner/TGBotApi"
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

	dbs := &requestHandler.DbStoragersPool{In: dbsIn, WorkersNumber: workers_n}
	dbs.Run()

	cache := requestHandler.NewMemoryCache()
	imgPath := s.config["chatscanner"]["images_path"].(string)
	os.MkdirAll(imgPath, os.ModePerm);
	context := requestHandler.AppContext{
		Db:               db,
		DownloadRequests: dr,
		PoolStop:         poolStoper,
		BotApi:           botApi,
		CfApi:            clApi,
		Cache:            cache,
		Logger:           s.logger,
		ImagesPath:       imgPath,
	}

	s.rAPIHandler.SetAppContext(&context)
	s.rAPIHandler.RegisterHandlers()

	s.srv = &http.Server{Handler: s.rAPIHandler}

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(poolStoper)
		l, err := net.Listen("unix", s.config["server"]["socket"].(string))
		if err != nil {
			s.logger.Println(err)
			wg.Done()
		}

		if err := os.Chmod(s.config["server"]["socket"].(string), 0777); err != nil {
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
