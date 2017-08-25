package requestHandler

import (
	"fmt"
	"github.com/zwirec/TGChatScanner/models"
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"sync"
	"strings"
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
			defer wg.Done()
			dsp.runStorager()
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
			Src:    strings.TrimPrefix(in.LocalPath, appContext.ImagesPath),
			ChatID: in.Basics.From,
			Date:   in.Basics.Sent,
		}
		var tags []models.Tag
		for _, t := range in.Basics.Tags {
			tags = append(tags, models.Tag{Name: t})
		}
		tx := appContext.DB.Begin()
		if err := img.CreateImageWithTags(tx, tags); err != nil {
			tx.Rollback()
			err = fmt.Errorf("failed on storing image: %s", err)
			NonBlockingNotify(in.Basics.Errorc, err)
		} else {
			select {
			case <-in.Basics.BasicContext.Done():
				tx.Rollback()
			default:
				NonBlockingNotify(in.Basics.Errorc, nil)
				tx.Commit()
			}
		}
	}
}
