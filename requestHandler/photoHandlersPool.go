package requestHandler

import (
    "log"
    "bytes"
    "github.com/zwirec/TGChatScanner/TGBotApi"
    "sync"
    "github.com/zwirec/TGChatScanner/clarifaiApi"
    "fmt"
)

type PhotoHandlersPool struct {
    photos        chan *PhotoHandleTask
    Results       chan *DownloadTask
    workersNumber int
    wg            sync.WaitGroup
}

type Photo struct {
    PhotoSize TGBotApi.PhotoSize
    From      TGBotApi.Chat
    Date      int
}

type PhotoHandleTask struct {
    Data    *Photo
    handler PhotoHandler
}

type PhotoHandler func(url string) ([]string, error)

func createLocalFilePath(fileId string) string {
    var buff bytes.Buffer
    buff.WriteString(fileId)
    buff.WriteString("/")
    buff.WriteString(fileId)
    return buff.String()
}

func NewPhotoHandlersPool(workersNumber int, queueSize int) *PhotoHandlersPool {
    tasks := make(chan *PhotoHandleTask, queueSize)
    results := make(chan *DownloadTask, queueSize)
    ph := &PhotoHandlersPool{
        photos:        tasks,
        workersNumber: workersNumber,
        Results:       results,
    }
    return ph
}

func (p *PhotoHandlersPool) RequestPhotoHandling(message *TGBotApi.Message, api *clarifaiApi.ClarifaiApi) {
    photo := &Photo{
        PhotoSize: message.Photo[len(message.Photo)-1],
        From:      message.Chat,
        Date:      message.Date,
    }
    handler := func(url string) ([]string, error) {
        return api.RecognizeImage(url, 0.9)
    }
    p.photos <- &PhotoHandleTask{photo, handler}
}

func (p *PhotoHandlersPool) Stop() {
    close(p.photos)
    p.wg.Wait()
}

func (p *PhotoHandlersPool) runHandler(tasks chan *PhotoHandleTask) {
    for task := range tasks {
        fileInfo, err := TGBotApi.PrepareFile(task.Data.PhotoSize.FileId)
        if err != nil {
            log.Printf("unable to download %s: %s", fileInfo.FileId, err)
            return
        }
        downloadRequest := FileDownloadRequest{fileInfo.FilePath, createLocalFilePath(fileInfo.FileId)}
        downloadTask := &DownloadTask{
            wg:       sync.WaitGroup{},
            FileInfo: downloadRequest,
            Result:   false,
        }
        p.Results <- downloadTask

        tags, err := task.handler(fileInfo.FilePath)

        if err != nil {
            log.Printf("recognition failed on %d: %s", task.Data.PhotoSize.FileId, err)
            //TODO: store spectial tag for this kind of images
            return
        }
        downloadTask.wg.Wait()
        if !downloadTask.Result {
            log.Printf("unable to store tags on %d: %s", fileInfo.FileId, err)
            return
        }
        //TODO: store file in db with tags
        fmt.Println(tags)
    }
    p.wg.Done()
}

func (p PhotoHandlersPool) Run() {
    for i := 0; i < p.workersNumber; i++ {
        p.wg.Add(1)
        go p.runHandler(p.photos)
    }
    p.wg.Done()
}
