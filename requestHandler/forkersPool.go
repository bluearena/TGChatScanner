package requestHandler

import "sync"

type FileInfo struct {
	FileId  string
	FileUrl string
}

type ForkersPool struct {
	In            chan *PreparedFile
	Out1          chan *FileLink
	Out2          chan *FileInfo
	Done          chan struct{}
	Caster1       ToOut1
	Caster2       ToOut2
	WorkersNumber int
}

type ToOut1 func(*PreparedFile) (*FileLink, error)

type ToOut2 func(result *PreparedFile) (*FileInfo, error)

func (fp *ForkersPool) Run(out1queue int, out2queue int) (out1 chan *FileLink, out2 chan *FileInfo) {
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
	go func() {
		wg.Wait()
		close(fp.Out1)
		close(fp.Out2)
	}()
	return fp.Out1, fp.Out2
}

func (fp *ForkersPool) fork() {
	for in := range fp.In {
		out1, err1 := fp.Caster1(in)
		out2, err2 := fp.Caster2(in)
		if err1 == nil && err2 == nil {
			select {
			case fp.Out1 <- out1:
			case fp.Out2 <- out2:
			case <-fp.Done:
				return
			}
		} else if err1 != nil && err2 == nil{
			select {
			case fp.Out2 <- out2:
			case <-fp.Done:
				return
			}
		}else if err1 == nil && err2 != nil{
			select {
			case fp.Out1 <- out1:
			case <-fp.Done:
				return
			}
		}
	}
}
