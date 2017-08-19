package requestHandler

import (
	"sync"
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
	Context map[string]interface{}
}

type FilePreparatorsPool struct {
	In            chan *FileBasic
	Out           chan *PreparedFile
	Done          chan struct{}
	WorkersNumber int
}

func (fpp *FilePreparatorsPool) Run(outBufferSize int) (chan *PreparedFile) {
	fpp.Out = make(chan *PreparedFile, outBufferSize)
	var wg sync.WaitGroup
	wg.Add(fpp.WorkersNumber)
	for i := 0; i < fpp.WorkersNumber; i++ {
		go func() {
			preparatorWorker(fpp.In, fpp.Out, fpp.Done)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(fpp.Out)
	}()
	return fpp.Out
}

func preparatorWorker(toPrepare chan *FileBasic, result chan *PreparedFile, done chan struct{}) {
	for in := range toPrepare {
		fileId := in.FileId
		file, err := appContext.BotApi.PrepareFile(fileId)
		fl := FileLink{
			FileDowloadUrl: appContext.BotApi.EncodeDownloadUrl(file.FilePath),
			LocalPath:      appContext.ImagesPath,
			Basics:         *in,
		}
		fpResult := &PreparedFile{fl, err}
		select {
		case result <- fpResult:
		case <-done:
			return
		}
	}
}
