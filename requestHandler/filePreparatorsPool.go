package requestHandler

import (
	"bytes"
	"fmt"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"sync"
)

type FilePreparationsPool struct {
	In            chan *file.FileBasic
	Out           chan *file.PreparedFile
	Done          chan struct{}
	WorkersNumber int
}

func (fpp *FilePreparationsPool) Run(outBufferSize int, finished *sync.WaitGroup) chan *file.PreparedFile {
	fpp.Out = make(chan *file.PreparedFile, outBufferSize)
	var wg sync.WaitGroup

	wg.Add(fpp.WorkersNumber)
	for i := 0; i < fpp.WorkersNumber; i++ {
		go func() {
			defer wg.Done()
			preparationWorker(fpp.In, fpp.Out, fpp.Done)
		}()
	}
	finished.Add(1)
	go func() {
		wg.Wait()
		close(fpp.Out)
		finished.Done()
	}()
	return fpp.Out
}

func preparationWorker(toPrepare chan *file.FileBasic, result chan *file.PreparedFile, done chan struct{}) {
	for in := range toPrepare {
		fileId := in.FileId
		f, err := appContext.BotAPI.PrepareFile(fileId)
		if err != nil {
			err = fmt.Errorf("error during preparation stage on %s: %s", in.FileId, err)
			NonBlockingNotify(in.Errorc, err)
			continue
		}

		url, err := appContext.BotAPI.EncodeDownloadURL(f.FilePath)
		if err != nil {
			err = fmt.Errorf("incorrect url during preparation stage on %s: %s", in.FileId, err)
			NonBlockingNotify(in.Errorc, err)
			continue
		}
		status := file.Undefiend
		fl := &file.FileLink{
			FileDownloadURL: url,
			LocalPath:       BuildLocalPath(fileId),
			Basics:          in,
			Status:          &status,
		}
		fpResult := &file.PreparedFile{Link: fl}
		select {
		case <-in.BasicContext.Done():
			continue
		case result <- fpResult:
		case <-done:
			return
		}
	}
}

func BuildLocalPath(fileId string) string {
	var buff bytes.Buffer
	buff.WriteString(appContext.ImagesPath)
	buff.WriteString("/")
	buff.WriteString(fileId)
	return buff.String()
}

func NonBlockingNotify(errorc chan error, err error) {
	select {
	case errorc <- err:
	default:
		return
	}
}
