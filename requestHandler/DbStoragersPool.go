package requestHandler

import (
	"github.com/zwirec/TGChatScanner/models"
	"sync"
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
		img := &models.Image{
			Src:    in.LocalPath,
			ChatID: in.Basics.Context["from"].(int64),
		}
		tags := in.Basics.Context["tags"].([]string)
		if err := img.CreateImageWithTags(appContext.Db, tags); err != nil {
			appContext.SysLogger.Printf("failed on storaging image: %s", err)
		}
	}
}
