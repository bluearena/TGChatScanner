package requestHandler

import (
	"errors"
	"fmt"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"github.com/zwirec/TGChatScanner/requestHandler/recognizers"
	"sync/atomic"
)

var (
	ErrWrongForkPath    = errors.New("wrong fork path")
	ErrBadDeforkAttempt = errors.New("bad defork attempt")
)

func CastToFileLink(pf *file.PreparedFile) (*file.FileLink, error) {
	if pf.Error != nil {
		err := fmt.Errorf("invalid prepared file on fork: %s", pf.Error)
		NonBlockingNotify(pf.Link.Basics.Errorc, err)
		return nil, pf.Error
	}
	return pf.Link, nil
}

func CastToFileInfo(pf *file.PreparedFile) (*file.FileInfo, error) {
	if pf.Link.Basics.Type != "photo" {
		err := fmt.Errorf("invalid file type on fork: %s", pf.Link.Basics.Type)
		NonBlockingNotify(pf.Link.Basics.Errorc, err)
		return nil, ErrWrongForkPath
	}
	return (*file.FileInfo)(pf.Link), nil
}

func CastFromDownloadedFile(df *file.DownloadedFile) (*file.CompleteFile, error) {
	if df.Error != nil {
		err := fmt.Errorf("invalid downloaded file on defork: %s", df.Error)
		atomic.StoreInt32(df.Link.Status, file.DownloadFailed)
		NonBlockingNotify(df.Link.Basics.Errorc, err)
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
		err := fmt.Errorf("invalid recognized photo on defork: %s", rp.Error)
		atomic.StoreInt32(rp.Link.Status, file.RecFailed)
		NonBlockingNotify(rp.Link.Basics.Errorc, err)
		return nil, rp.Error
	}

	status := atomic.SwapInt32(rp.Link.Status, file.RecSuccess)
	if status == file.DownloadSuccess {
		return (*file.CompleteFile)(rp.Link), nil
	}
	return nil, ErrBadDeforkAttempt
}
