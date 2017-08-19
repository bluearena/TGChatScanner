package requestHandler

import (
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type DownloadTask struct {
	wg       sync.WaitGroup
	FileInfo FileDownloadRequest
	Result   bool
}

type FileDownloadRequest struct {
	Url       string
	LocalPath string
}

type FileDownloadersPool struct {
	wg            sync.WaitGroup
	tasks         chan *DownloadTask
	WorkersNumber int
}

func NewFileDownloaderPool(workersNumber int, queueSize int) *FileDownloadersPool {
	tasks := make(chan *DownloadTask, queueSize)
	pool := &FileDownloadersPool{tasks: tasks, WorkersNumber: workersNumber}
	return pool
}

func (fd *FileDownloadersPool) RequestDownloading(task *DownloadTask) {
	fd.tasks <- task
}

func (fd *FileDownloadersPool) Stop() {
	close(fd.tasks)
	fd.wg.Wait()
}

func (fd *FileDownloadersPool) Run() {
	for i := 0; i < fd.WorkersNumber; i++ {
		fd.wg.Add(1)
		go fd.runDownloader(fd.tasks)
	}
	fd.wg.Done()
}

func (p *FileDownloadersPool) runDownloader(tasks chan *DownloadTask) {
	for task := range tasks {
		defer task.wg.Done()
		downloadUrl := TGBotApi.EncodeDownloadUrl(task.FileInfo.Url)
		response, err := http.Get(downloadUrl)
		defer response.Body.Close()
		if err != nil {
			log.Printf("File download failed on %s: %s", task.FileInfo.Url, err)
			return
		}
		out, err := os.Create(task.FileInfo.LocalPath)
		if err != nil {
			log.Printf("File creation failed on %s: %s", task.FileInfo.LocalPath, err)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, response.Body)
		if err != nil {
			log.Printf("File write failed on %s: %s", task.FileInfo.LocalPath, err)
			return
		}
		task.Result = true
	}
	p.wg.Done()
}
