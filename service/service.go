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
    "github.com/zwirec/TGChatScanner/clarifaiApi"
    "github.com/zwirec/TGChatScanner/modelManager"
    "io/ioutil"
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

    usr, err := user.Current()

    if err != nil {
        s.logger.Println(err)
        return err
    }

    configPath := usr.HomeDir + "/.config/tgchatscanner/config.json"

	if err := s.parseConfig(configPath); err != nil {
        s.logger.Println(err)
        return err
    }

    s.signalProcessing()

    db, err := modelManager.ConnectToDB(s.config["db"])

    api := clarifaiApi.NewClarifaiApi(clarifaiApi.ApiKey)

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
        CfApi:         api,
        Cache:         &cache,
        Logger:        s.logger,
    }

    s.rAPIHandler.SetAppContext(&context)
    s.rAPIHandler.RegisterHandlers()

    s.srv = &http.Server{Addr: ":" + s.config["server"]["port"].(string), Handler: s.rAPIHandler}

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
    file, err := ioutil.ReadFile(filename)

    if err != nil {
        return err
    }

    if err = json.Unmarshal(file, &s.config); err != nil {
		return err
	}

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
