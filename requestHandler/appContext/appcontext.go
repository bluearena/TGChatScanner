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
	ImagesPath       string
	Hostname         string
}

type key string

var (
	MessageKey key = "message"
)

func SetAppContext(context *AppContext) {
	DB = context.DB
	DownloadRequests = context.DownloadRequests
	CfAPI = context.CfAPI
	BotAPI = context.BotAPI
	Cache = context.Cache
	ErrLogger = context.ErrLogger
	AccessLogger = context.ErrLogger
	ImagesPath = context.ImagesPath
	Hostname = context.Hostname
}
