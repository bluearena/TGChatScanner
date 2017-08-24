package requestHandler

import (
	"bytes"
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

func (fpp *FilePreparationsPool) Run(outBufferSize int, finished sync.WaitGroup) chan *file.PreparedFile {
	fpp.Out = make(chan *file.PreparedFile, outBufferSize)
	var wg sync.WaitGroup

	wg.Add(fpp.WorkersNumber)
	for i := 0; i < fpp.WorkersNumber; i++ {
		go func() {
			preparationWorker(fpp.In, fpp.Out, fpp.Done)
			wg.Done()
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
		f, err := appContext.BotApi.PrepareFile(fileId)
		if err != nil {
			appContext.ErrLogger.Printf("error during preparation stage on %s: %s", in.FileId, err)
			continue
		}
		fl := &file.FileLink{
			FileDownloadUrl: appContext.BotApi.EncodeDownloadUrl(f.FilePath),
			LocalPath:       BuildLocalPath(fileId),
			Basics:          in,
		}
		fpResult := &file.PreparedFile{Link: fl}
		select {
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
