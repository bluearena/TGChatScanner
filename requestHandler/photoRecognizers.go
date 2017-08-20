package requestHandler

import "sync"

type PhotoRecognizersPool struct {
	In            chan *FileInfo
	Out           chan *RecognizedPhoto
	Done          chan struct{}
	WorkersNumber int
}

type RecognizedPhoto struct {
	FileId string
	Tags   []string
	Error  error
}

func (frp *PhotoRecognizersPool) Run(queueSize int) chan *RecognizedPhoto {
	frp.Out = make(chan *RecognizedPhoto, queueSize)
	var wg sync.WaitGroup
	wg.Add(frp.WorkersNumber)
	for i := 0; i < frp.WorkersNumber; i++ {
		go func() {
			frp.runPhotoRecognizer()
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(frp.Out)
	}()
	return frp.Out
}

func (frp *PhotoRecognizersPool) runPhotoRecognizer() {
	for in := range frp.In {
		tags, err := appContext.CfApi.RecognizeImage(in.FileUrl, 0.9)
		rp := &RecognizedPhoto{in.FileId, tags, err}
		select {
		case frp.Out <- rp:
		case <-frp.Done:
			return
		}
	}
}
