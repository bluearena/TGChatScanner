package recognizers

import (
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	"sync"
	"github.com/zwirec/TGChatScanner/requestHandler/filetypes"
)

func (frp *PhotoRecognizersPool) Run(queueSize int, finished *sync.WaitGroup) chan *RecognizedPhoto {
	frp.Out = make(chan *RecognizedPhoto, queueSize)
	var wg sync.WaitGroup
	wg.Add(frp.WorkersNumber)
	for i := 0; i < frp.WorkersNumber; i++ {
		go func() {
			defer wg.Done()
			frp.runPhotoRecognizer()
		}()
	}
	finished.Add(1)
	go func() {
		wg.Wait()
		close(frp.Out)
		finished.Done()
	}()
	return frp.Out
}

func (frp *PhotoRecognizersPool) runPhotoRecognizer() {
	for in := range frp.In {

		appContext.ErrLogger.Printf("comes on rec: %+v", *in)
		tags, err := appContext.CfAPI.RecognizeImage(in.FileDownloadURL, 0.9)
		in.Basics.Tags = tags
		rp := &RecognizedPhoto{(*filetypes.FileLink)(in), err}
		appContext.ErrLogger.Printf("comes from rec: %+v", *rp)
		select {
		case frp.Out <- rp:
		case <-frp.Done:
			return
		}
	}
}
