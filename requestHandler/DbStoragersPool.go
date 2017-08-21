package requestHandler

import (
	"sync"
	"fmt"
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
		//TODO: Acctually store file in the db
		switch in.Basics.Type {
		case "photo":
			fmt.Errorf("Photo to db\n")
		case "chatPhoto":
			fmt.Errorf("ChatPhoto to db\n")
		}
	}
}
