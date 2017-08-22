package requestHandler

import (
	"errors"
	"github.com/patrickmn/go-cache"
)

var (
	ErrWrongForkPath    = errors.New("wrong fork path")
	ErrBadDeforkAttempt = errors.New("incomplete file")
)

func CastToFileLink(pf *PreparedFile) (*FileLink, error) {
	if pf.Error != nil {
		return nil, pf.Error
	}
	return pf.Link, nil
}

func CastToFileInfo(pf *PreparedFile) (*FileInfo, error) {
	if pf.Link.Basics.Type != "photo" {
		return nil, ErrWrongForkPath
	}
	return &FileInfo{pf.Link.Basics.FileId, pf.Link.FileDowloadUrl}, nil
}

func CastFromDownloadedFile(df *DownloadedFile) (*CompleteFile, error) {
	if df.Error != nil {
		appContext.SysLogger.Printf("invalid downloaded file on defork: %s", df.Error)
		return nil, df.Error
	}
	fID := df.Link.Basics.FileId
	exists := appContext.Cache.Add(fID, df.Link, cache.DefaultExpiration)
	if exists != nil {
		tags, _ := appContext.Cache.Get(fID)
		df.Link.Basics.Context["tags"] = tags
		link := (*CompleteFile)(df.Link)
		return link, nil
	}
	return nil, ErrBadDeforkAttempt
}

func CastFromRecognizedPhoto(rp *RecognizedPhoto) (*CompleteFile, error) {
	if rp.Error != nil {
		appContext.SysLogger.Printf("invalid recognized photo on defork: %s", rp.Error)
		return nil, rp.Error
	}
	fID := rp.FileId
	exists := appContext.Cache.Add(fID, rp.Tags, cache.DefaultExpiration)
	if exists != nil {
		lk, _ := appContext.Cache.Get(fID)
		link := lk.(*FileLink)
		link.Basics.Context["tags"] = rp.Tags
		return (*CompleteFile)(link), nil
	}
	return nil, ErrBadDeforkAttempt
}
