package requestHandler

import (
	"errors"
	"github.com/patrickmn/go-cache"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"github.com/zwirec/TGChatScanner/requestHandler/recognizers"
)

var (
	ErrWrongForkPath    = errors.New("wrong fork path")
	ErrBadDeforkAttempt = errors.New("incomplete file")
)

func CastToFileLink(pf *file.PreparedFile) (*file.FileLink, error) {
	if pf.Error != nil {
		return nil, pf.Error
	}
	return pf.Link, nil
}

func CastToFileInfo(pf *file.PreparedFile) (*file.FileInfo, error) {
	if pf.Link.Basics.Type != "photo" {
		return nil, ErrWrongForkPath
	}
	return &file.FileInfo{pf.Link.Basics.FileId, pf.Link.FileDownloadURL}, nil
}

func CastFromDownloadedFile(df *file.DownloadedFile) (*file.CompleteFile, error) {
	if df.Error != nil {
		appContext.ErrLogger.Printf("invalid downloaded file on defork: %s", df.Error)
		return nil, df.Error
	}
	fID := df.Link.Basics.FileId
	exists := appContext.Cache.Add(fID, df.Link, cache.DefaultExpiration)
	if exists != nil {
		tags, _ := appContext.Cache.Get(fID)
		df.Link.Basics.Context["tags"] = tags
		link := (*file.CompleteFile)(df.Link)
		return link, nil
	}
	return nil, ErrBadDeforkAttempt
}

func CastFromRecognizedPhoto(rp *recognizers.RecognizedPhoto) (*file.CompleteFile, error) {
	if rp.Error != nil {
		appContext.ErrLogger.Printf("invalid recognized photo on defork: %s", rp.Error)
		return nil, rp.Error
	}
	fID := rp.FileId
	exists := appContext.Cache.Add(fID, rp.Tags, cache.DefaultExpiration)
	if exists != nil {
		lk, _ := appContext.Cache.Get(fID)
		link := lk.(*file.FileLink)
		link.Basics.Context["tags"] = rp.Tags
		return (*file.CompleteFile)(link), nil
	}
	return nil, ErrBadDeforkAttempt
}
