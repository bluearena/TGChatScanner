package TGBotApi

import (
    "net/http"
    "log"
    "os"
    "io"
)

type DownloadPromise struct {
    Done     chan bool
    FileInfo FileDownloadRequest
}

type FileDownloadRequest struct {
    Url       string
    LocalPath string
}

type FileDownloaderPool struct {
    promises      chan DownloadPromise
    WorkersNumber int
}

func NewFileDownloaderPool(workersNumber int, queueSize int) *FileDownloaderPool {
    urls := make(chan DownloadPromise, queueSize)
    pool := &FileDownloaderPool{promises: urls, WorkersNumber: workersNumber}
    pool.init()
    return pool
}

func (fd *FileDownloaderPool) RequestDownloading(promise DownloadPromise){
    fd.promises <- promise
}

func (fd *FileDownloaderPool) init() {
    for p := range fd.promises {
        go fileDownloader(p)
    }
}

func fileDownloader(promise DownloadPromise) {
    downloadUrl := encodeDownloadUrl(promise.FileInfo.Url)
    response, err := http.Get(downloadUrl)
    defer response.Body.Close()
    if err != nil {
        log.Printf("File download failed on %s: %s", promise.FileInfo.Url, err)
        promise.Done <- false
        return
    }
    out, err := os.Create(promise.FileInfo.LocalPath)
    if err != nil {
        log.Printf("File creation failed on %s: %s", promise.FileInfo.LocalPath, err)
        promise.Done <- false
        return
    }
    defer out.Close()
    _, err = io.Copy(out, response.Body)
    if err != nil {
        log.Printf("File write failed on %s: %s", promise.FileInfo.LocalPath, err)
        promise.Done <- false
        return
    }
    //TODO: Save file to database
    promise.Done <- true
    close(promise.Done)
}

