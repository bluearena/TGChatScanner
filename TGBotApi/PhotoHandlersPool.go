package TGBotApi

import "log"

type PhotoHandlersPool struct {
    photos        chan PhotoSize
    workersNumber int
}

func NewPhotoHandlersPool(workersNumber int, queueSize int) *PhotoHandlersPool {
    photos := make(chan PhotoSize, queueSize)
    ph := &PhotoHandlersPool{photos: photos,
        workersNumber: workersNumber}
    ph.init()
    return ph
}

func (p *PhotoHandlersPool) RequestPhotoHandling(photo PhotoSize) {
    p.photos <- photo
}

func photoHandler(photos chan PhotoSize) {
    for photo := range photos {
        err := handlePhotoSize(photo)
        if err != nil {
            log.Printf("photoHandler failed on %d: %s", photo.FileId, err)
        }
    }
}

func (p PhotoHandlersPool) init() {
    for i := 0; i < p.workersNumber; i++ {
        go photoHandler(p.photos)
    }
}

//TODO: Implement it
func handlePhotoSize(photo PhotoSize) error {
    return nil
}