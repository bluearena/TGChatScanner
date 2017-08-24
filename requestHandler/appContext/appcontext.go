package appContext

import (
	"github.com/jinzhu/gorm"
	memcache "github.com/patrickmn/go-cache"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"github.com/zwirec/TGChatScanner/clarifaiApi"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"log"
)

var (
	DB               *gorm.DB
	DownloadRequests chan *file.FileBasic
	CfApi            *clarifaiApi.ClarifaiApi
	BotApi           *TGBotApi.BotApi
	Cache            *memcache.Cache
	ErrLogger        *log.Logger
	AccessLogger     *log.Logger
	ImagesPath       string
	Hostname         string
)

type AppContext struct {
	DB               *gorm.DB
	DownloadRequests chan *file.FileBasic
	CfApi            *clarifaiApi.ClarifaiApi
	BotApi           *TGBotApi.BotApi
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
	CfApi = context.CfApi
	BotApi = context.BotApi
	Cache = context.Cache
	ErrLogger = context.ErrLogger
	AccessLogger = context.ErrLogger
	ImagesPath = context.ImagesPath
	Hostname = context.Hostname
}
