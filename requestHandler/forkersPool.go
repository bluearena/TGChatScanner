package requestHandler

import "sync"

type FileInfo struct {
	FileId  string
	FileUrl string
}

type ForkersPool struct {
	In             chan *PreparedFile
	Out1           chan *FileLink
	Out2           chan *FileInfo
	Done           chan struct{}
	ForkToFileLink InToFileLink
	ForkToFileInfo InToFileInfo
	WorkersNumber  int
}

type InToFileLink func(*PreparedFile) (*FileLink, error)

type InToFileInfo func(result *PreparedFile) (*FileInfo, error)

func (fp *ForkersPool) Run(out1queue int, out2queue int, finished sync.WaitGroup) (out1 chan *FileLink, out2 chan *FileInfo) {
	fp.Out1 = make(chan *FileLink, out1queue)
	fp.Out2 = make(chan *FileInfo, out2queue)
	var wg sync.WaitGroup

	wg.Add(fp.WorkersNumber)
	for i := 0; i < fp.WorkersNumber; i++ {
		go func() {
			fp.fork()
			wg.Done()
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
			var wg sync.WaitGroup
			wg.Add(1)
			go func(outCopy *FileLink) {
				defer wg.Done()
				select {
				case fp.Out1 <- out1:
				case <-fp.Done:
					return
				}
			}(out1)

			select {
			case fp.Out2 <- out2:
			case <-fp.Done:

				return
			}

			wg.Wait()
		} else if err1 != nil && err2 == nil {
			select {
			case fp.Out2 <- out2:
			case <-fp.Done:
				return
			}
		} else if err1 == nil && err2 != nil {
			select {
			case fp.Out1 <- out1:
			case <-fp.Done:
				return
			}
		}
	}
}
