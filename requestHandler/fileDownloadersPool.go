package requestHandler

import (
	"sync"
	"os"
	"net/http"
	"io"
)

type DownloadedFile PreparedFile

type FileDownloadersPool struct {
	In            chan *FileLink
	Out           chan *DownloadedFile
	Done          chan struct{}
	WorkersNumber int
}

func (fdp *FileDownloadersPool) Run(queueSize int, finished sync.WaitGroup) (chan *DownloadedFile) {
	fdp.Out = make(chan *DownloadedFile, queueSize)
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
		err := downloadFile(in.FileDowloadUrl, in.LocalPath)
		df := &DownloadedFile{*in, err}
		select {
		case fdp.Out <- df:
		case <-fdp.Done:
			return
		}
	}
}

func downloadFile(url string, localPath string) error {
	resp, err := http.Get(url)
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
