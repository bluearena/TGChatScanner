package requestHandler

import (
	"errors"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"github.com/zwirec/TGChatScanner/requestHandler/recognizers"
	"sync/atomic"
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
	return (*file.FileInfo)(pf.Link), nil
}

func CastFromDownloadedFile(df *file.DownloadedFile) (*file.CompleteFile, error) {
	if df.Error != nil {
		appContext.ErrLogger.Printf("invalid downloaded file on defork: %s", df.Error)
		atomic.StoreInt32(df.Link.Status, file.DownloadFailed)
		return nil, df.Error
	}

	status := atomic.SwapInt32(df.Link.Status, file.DownloadSuccess)
	if status == file.RecSuccess {
		return (*file.CompleteFile)(df.Link), nil
	}
	return nil, ErrBadDeforkAttempt
}

func CastFromRecognizedPhoto(rp *recognizers.RecognizedPhoto) (*file.CompleteFile, error) {
	if rp.Error != nil {
		appContext.ErrLogger.Printf("invalid recognized photo on defork: %s", rp.Error)
		atomic.StoreInt32(rp.Link.Status, file.RecFailed)
		return nil, rp.Error
	}

	status := atomic.SwapInt32(rp.Link.Status, file.RecSuccess)

	if status == file.DownloadSuccess {
		return (*file.CompleteFile)(rp.Link), nil
	}
	return nil, ErrBadDeforkAttempt
}
