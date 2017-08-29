package appContext

import (
	"log"

	"github.com/jinzhu/gorm"
	memcache "github.com/patrickmn/go-cache"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/clarifaiApi"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
)

var (
	DB               *gorm.DB
	DownloadRequests chan *file.FileBasic
	CfAPI            *clarifaiAPI.ClarifaiAPI
	BotAPI           *TGBotAPI.BotAPI
	Cache            *memcache.Cache
	ErrLogger        *log.Logger
	AccessLogger     *log.Logger
	ImagesPath       string
	ImagesPrefix     string
	Hostname         string
)

type AppContext struct {
	DB               *gorm.DB
	DownloadRequests chan *file.FileBasic
	CfAPI            *clarifaiAPI.ClarifaiAPI
	BotAPI           *TGBotAPI.BotAPI
	Cache            *memcache.Cache
	ErrLogger        *log.Logger
	AccessLogger     *log.Logger
	ImagesPrefix     string
	ImagesPath       string
	Hostname         string
}

func SetAppContext(context *AppContext) {
	DB = context.DB
	DownloadRequests = context.DownloadRequests
	CfAPI = context.CfAPI
	BotAPI = context.BotAPI
	Cache = context.Cache
	ErrLogger = context.ErrLogger
	AccessLogger = context.AccessLogger
	ImagesPrefix = context.ImagesPrefix
	ImagesPath = context.ImagesPath
	Hostname = context.Hostname
}
