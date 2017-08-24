package requestHandler

import (
	"github.com/zwirec/TGChatScanner/models"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"sync"
)

type DbStoragesPool struct {
	In            chan *file.CompleteFile
	WorkersNumber int
}

func (dsp *DbStoragesPool) Run(finished *sync.WaitGroup) {
	var wg sync.WaitGroup
	wg.Add(dsp.WorkersNumber)
	for i := 0; i < dsp.WorkersNumber; i++ {
		go func() {
			dsp.runStorager()
			wg.Done()
		}()
	}
	finished.Add(1)
	go func() {
		wg.Wait()
		finished.Done()
	}()
}

func (dsp *DbStoragesPool) runStorager() {
	for in := range dsp.In {
		img := &models.Image{
			Src:    in.LocalPath,
			ChatID: in.Basics.Context["from"].(int64),
		}
		tags := in.Basics.Context["tags"].([]string)
		if err := img.CreateImageWithTags(appContext.DB, tags); err != nil {
			appContext.ErrLogger.Printf("failed on storing image: %s", err)
		}
	}
}