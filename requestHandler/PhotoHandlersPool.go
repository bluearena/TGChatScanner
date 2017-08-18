package requestHandler

import (
    "log"
    "bytes"
    "github.com/zwirec/TGChatScanner/TGBotApi"
)

type PhotoHandlersPool struct {
    photos        chan TGBotApi.PhotoSize
    workersNumber int
}

func createLocalFilePath(fileId string) string {
    var buff bytes.Buffer
    buff.WriteString(fileId)
    buff.WriteString("/")
    buff.WriteString(fileId)
    return buff.String()
}

func NewPhotoHandlersPool(workersNumber int, queueSize int, results *FileDownloaderPool) *PhotoHandlersPool {
    photos := make(chan TGBotApi.PhotoSize, queueSize)

    ph := &PhotoHandlersPool{photos: photos,
        workersNumber: workersNumber}
    ph.init(results)
    return ph
}

func (p *PhotoHandlersPool) RequestPhotoHandling(photo TGBotApi.PhotoSize) {
    p.photos <- photo
}
func (p *PhotoHandlersPool) Stop() {
    close(p.photos)
}
func photoHandler(photos chan TGBotApi.PhotoSize, resultHandlers *FileDownloaderPool) {
    for photo := range photos {
        fileInfo, err := TGBotApi.PrepareFile(photo.FileId)
        if err != nil {
            log.Printf("unable to download %s: %s", fileInfo.FileId, err)
            return
        }

        done := make(chan bool, 1)
        downloadRequest := FileDownloadRequest{fileInfo.FilePath, createLocalFilePath(fileInfo.FileId)}
        promise := DownloadPromise{done, downloadRequest}
        resultHandlers.RequestDownloading(promise)

        err = handlePhotoSize(photo)
        if err != nil {
            log.Printf("recognition failed on %d: %s", photo.FileId, err)
            //TODO: store spectial tag for this kind of images
            return
        }

        isFileReady := <-promise.Done
        if !isFileReady {
            log.Printf("unable to store tags on %d: %s", fileInfo.FileId, err)
            return
        }
        //TODO: store tags in db
    }
}

func (p PhotoHandlersPool) init(resultHandlers *FileDownloaderPool) {
    for i := 0; i < p.workersNumber; i++ {
        go photoHandler(p.photos, resultHandlers)
    }
}

//TODO: Implement it
func handlePhotoSize(photo TGBotApi.PhotoSize) error {
    return nil
}
