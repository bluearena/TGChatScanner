package deforkers

import (
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"sync"
)

func (dp *DeforkersPool) Run(queueSize int, finished *sync.WaitGroup) chan *file.CompleteFile {
	dp.Out = make(chan *file.CompleteFile, queueSize)
	var wg sync.WaitGroup
	wg.Add(dp.WorkersNumber)
	for i := 0; i < dp.WorkersNumber; i++ {
		go func() {
			dp.defork()
			wg.Done()
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
			appContext.SysLogger.Printf("comes on defork1: %+v", *in1)
			out, err := dp.DeforkDownloaded(in1)
			if err != nil {
				continue
			}

			appContext.SysLogger.Printf("comes from defork1: %+v", *out)
			select {
			case dp.Out <- out:
			case <-dp.Done:
				return
			}
		case in2 := <-dp.In2:
			appContext.SysLogger.Printf("comes on defork2: %+v", *in2)
			out, err := dp.DeforkRecognized(in2)
			if err != nil {
				continue
			}

			appContext.SysLogger.Printf("comes from defork2: %+v", *out)
			select {
			case dp.Out <- out:
			case <-dp.Done:
				return
			}
		}
	}
}
