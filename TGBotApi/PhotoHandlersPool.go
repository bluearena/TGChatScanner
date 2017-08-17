package TGBotApi

import (
    "log"
    "bytes"
)

type PhotoHandlersPool struct {
    photos        chan PhotoSize
    workersNumber int
}

func createLocalFilePath(fileId string) string {
    var buff bytes.Buffer
    buff.WriteString(fileId)
    buff.WriteString("/")
    buff.WriteString(fileId)
    return buff.String()
}

func NewPhotoHandlersPool(workersNumber int, queueSize int, results FileDownloaderPool) *PhotoHandlersPool {
    photos := make(chan PhotoSize, queueSize)

    ph := &PhotoHandlersPool{photos: photos,
        workersNumber: workersNumber}
    ph.init(results)
    return ph
}

func (p *PhotoHandlersPool) RequestPhotoHandling(photo PhotoSize) {
    p.photos <- photo
}

func photoHandler(photos chan PhotoSize, resultHandlers FileDownloaderPool) {
    for photo := range photos {
        fileInfo, err := PrepareFile(photo.FileId)
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

func (p PhotoHandlersPool) init(resultHandlers FileDownloaderPool) {
    for i := 0; i < p.workersNumber; i++ {
        go photoHandler(p.photos, resultHandlers)
    }
}

//TODO: Implement it
func handlePhotoSize(photo PhotoSize) error {
    return nil
}
