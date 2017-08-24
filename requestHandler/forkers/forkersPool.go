package forkers

import (
	"github.com/zwirec/TGChatScanner/requestHandler/appContext"
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"sync"
)

func (fp *ForkersPool) Run(out1queue int, out2queue int, finished *sync.WaitGroup) (out1 chan *file.FileLink, out2 chan *file.FileInfo) {
	fp.Out1 = make(chan *file.FileLink, out1queue)
	fp.Out2 = make(chan *file.FileInfo, out2queue)
	var wg sync.WaitGroup

	wg.Add(fp.WorkersNumber)
	for i := 0; i < fp.WorkersNumber; i++ {
		go func() {
			defer wg.Done()
			fp.fork()
		}()
	}
	finished.Add(1)
	go func() {
		wg.Wait()
		close(fp.Out1)
		close(fp.Out2)
		finished.Done()
	}()
	return fp.Out1, fp.Out2
}

func (fp *ForkersPool) fork() {
	for in := range fp.In {
		if in.Error != nil {
			continue
		}
		out1, err1 := fp.ForkToFileLink(in)
		out2, err2 := fp.ForkToFileInfo(in)
		if err1 == nil && err2 == nil {

			select {
			case <-in.Link.Basics.BasicContext.Done():
				continue
			case fp.Out1 <- out1:

				select {
				case fp.Out2 <- out2:
				case <-in.Link.Basics.BasicContext.Done():
					continue
				case <-fp.Done:
					return
				}

			case fp.Out2 <- out2:

				select {
				case fp.Out1 <- out1:
				case <-in.Link.Basics.BasicContext.Done():
					continue
				case <-fp.Done:
					return
				}
			case <-fp.Done:
				return
			}

		} else if err1 != nil && err2 == nil {

			select {
			case <-in.Link.Basics.BasicContext.Done():
				continue
			case fp.Out2 <- out2:
			case <-fp.Done:
				return
			}

		} else if err1 == nil && err2 != nil {
			select {
			case <-in.Link.Basics.BasicContext.Done():
				continue
			case fp.Out1 <- out1:
			case <-fp.Done:
				return
			}

		}
	}
}
