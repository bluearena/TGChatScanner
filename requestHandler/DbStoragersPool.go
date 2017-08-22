package requestHandler

import (
	"sync"
	"github.com/zwirec/TGChatScanner/models"
)

type DbStoragersPool struct {
	In            chan *CompleteFile
	WorkersNumber int
}

func (dsp *DbStoragersPool) Run(finished sync.WaitGroup) {
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

func (dsp *DbStoragersPool) runStorager() {
	for in := range dsp.In {
		appContext.Logger.Printf("Comes to db: %+v", *in)
		img := &models.Image{
			Src:    in.LocalPath,
			ChatID: in.Basics.Context["from"].(uint64),
		}
		tags := in.Basics.Context["tags"].([]string)
		if err := img.CreateImageWithTags(appContext.Db, tags);
			err != nil {
				appContext.Logger.Printf("failed on storaging image: %s", err)
		}
	}
}
