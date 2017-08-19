package requestHandler

import (
	"errors"
)

var (
	ErrWrongForkPath    = errors.New("wrong fork path")
	ErrBadDeforkAttempt = errors.New("incomplete file")
)

func CastToFileLink(pf *PreparedFile) (*FileLink, error) {
	return &pf.Link, nil
}

func CastToFileInfo(pf *PreparedFile) (*FileInfo, error) {
	if pf.Link.Basics.Type != "photo" {
		return nil, ErrWrongForkPath
	}
	return &FileInfo{pf.Link.Basics.FileId, pf.Link.FileDowloadUrl}, nil
}

func CastFromDownloadedFile(df *DownloadedFile) (*CompleteFile, error) {
	fID := df.Link.Basics.FileId
	ok := appContext.Cache.Add(fID, &df.Link)
	if !ok {
		tags, _ := appContext.Cache.Get(fID)
		df.Link.Basics.Context["tags"] = tags.([]string)
		link := CompleteFile(df.Link)
		return &link, nil
	}
	return nil, ErrBadDeforkAttempt
}

func CastFromRecognizedPhoto(rp *RecognizedPhoto) (*CompleteFile, error) {
	fID := rp.FileId
	ok := appContext.Cache.Add(fID, rp.Tags)
	if !ok {
		lk, _ := appContext.Cache.Get(fID)
		link := lk.(*CompleteFile)
		link.Basics.Context["tags"] = rp.Tags
		return link, nil
	}
	return nil, ErrBadDeforkAttempt
}
