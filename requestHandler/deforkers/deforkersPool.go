package deforkers

import (
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"sync"
)

func (dp *DeforkersPool) Run(queueSize int, finished *sync.WaitGroup) chan *file.CompleteFile {
	dp.Out = make(chan *file.CompleteFile, queueSize)
	var wg sync.WaitGroup
	wg.Add(dp.WorkersNumber)
	for i := 0; i < dp.WorkersNumber; i++ {
		go func() {
			defer wg.Done()
			dp.defork()
		}()
	}
	finished.Add(1)
	go func() {
		wg.Wait()
		close(dp.Out)
		finished.Done()
	}()
	return dp.Out
}

func (dp *DeforkersPool) defork() {
	for {
		select {
		case in1 := <-dp.In1:
			if in1 == nil {
				return
			}
			appContext.ErrLogger.Printf("comes on defork1: %+v\n%+v", *in1.Link,*in1.Link.Basics)
			out, err := dp.DeforkDownloaded(in1)
			if err != nil {
				continue
			}

			appContext.ErrLogger.Printf("comes from defork1: %+v\n%+v", *out, *out.Basics)
			select {
			case dp.Out <- out:
			case <-dp.Done:
				return
			}
		case in2 := <-dp.In2:
			if in2 == nil {
				return
			}
			appContext.ErrLogger.Printf("comes on defork2: %+v\n%+v", *in2.Link,*in2.Link.Basics)
			out, err := dp.DeforkRecognized(in2)
			if err != nil {
				continue
			}
			appContext.ErrLogger.Printf("comes from defork2: %+v\n%+v", *out,*out.Basics)
			select {
			case dp.Out <- out:
			case <-dp.Done:
				return
			}
		}
	}
}
