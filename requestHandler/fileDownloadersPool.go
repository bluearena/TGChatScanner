package requestHandler

import (
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"io"
	"net/http"
	"os"
	"sync"
)

type FileDownloadersPool struct {
	In            chan *file.FileLink
	Out           chan *file.DownloadedFile
	Done          chan struct{}
	WorkersNumber int
}

func (fdp *FileDownloadersPool) Run(queueSize int, finished *sync.WaitGroup) chan *file.DownloadedFile {
	fdp.Out = make(chan *file.DownloadedFile, queueSize)
	var wg sync.WaitGroup
	wg.Add(fdp.WorkersNumber)

	for i := 0; i < fdp.WorkersNumber; i++ {
		go func() {
			fdp.runDownloader()
			wg.Done()
		}()
	}
	finished.Add(1)
	go func() {
		wg.Wait()
		close(fdp.Out)
		finished.Done()
	}()
	return fdp.Out
}

func (fdp *FileDownloadersPool) runDownloader() {
	for in := range fdp.In {
		err := downloadFile(in.FileDownloadURL, in.LocalPath)
		df := &file.DownloadedFile{in, err}
		select {
		case fdp.Out <- df:
		case <-fdp.Done:
			return
		}
	}
}

func downloadFile(URL string, localPath string) error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(localPath)
	}

	return err
}
