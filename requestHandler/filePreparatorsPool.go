package requestHandler

import (
	"bytes"
	"sync"
	"time"
)

type FileLink struct {
	FileDowloadUrl string
	LocalPath      string
	Basics         FileBasic
}

type PreparedFile struct {
	Link  FileLink
	Error error
}

type FileBasic struct {
	FileId  string
	Type    string
	Sent    time.Time
	Context map[string]interface{}
}

type FilePreparatorsPool struct {
	In            chan *FileBasic
	Out           chan *PreparedFile
	Done          chan struct{}
	WorkersNumber int
}

func (fpp *FilePreparatorsPool) Run(outBufferSize int, finished sync.WaitGroup) chan *PreparedFile {
	fpp.Out = make(chan *PreparedFile, outBufferSize)
	var wg sync.WaitGroup

	wg.Add(fpp.WorkersNumber)
	for i := 0; i < fpp.WorkersNumber; i++ {
		go func() {
			preparatorWorker(fpp.In, fpp.Out, fpp.Done)
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

func preparatorWorker(toPrepare chan *FileBasic, result chan *PreparedFile, done chan struct{}) {
	for in := range toPrepare {
		fileId := in.FileId
		file, err := appContext.BotApi.PrepareFile(fileId)
		if err != nil {
			appContext.Logger.Printf("error during preparation stage on %s: %s", in.FileId, err)
			continue
		}
		fl := FileLink{
			FileDowloadUrl: appContext.BotApi.EncodeDownloadUrl(file.FilePath),
			LocalPath:      BuildLocalPath(fileId),
			Basics:         *in,
		}
		fpResult := &PreparedFile{fl, nil}
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
