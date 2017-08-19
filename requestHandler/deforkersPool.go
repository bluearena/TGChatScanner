package requestHandler

import "sync"

type CompleteFile FileLink

type DeforkersPool struct {
	In1 chan *DownloadedFile
	In2 chan *RecognizedPhoto
	Out chan *CompleteFile
	In1Caster FromIn1
	In2Caster FromIn2
	Done chan struct{}
	WorkersNumber int
}

type FromIn1 func(*DownloadedFile)(*CompleteFile, error)

type FromIn2 func(*RecognizedPhoto)(*CompleteFile, error)

func (dp *DeforkersPool) Run(queueSize int) chan *CompleteFile{
	dp.Out = make(chan *CompleteFile, queueSize)
	var wg sync.WaitGroup
	wg.Add(dp.WorkersNumber)
	for i:= 0; i < dp.WorkersNumber; i++{
		go func() {
			dp.defork()
			wg.Done()
		}()
	}
	go func(){
		wg.Wait()
		close(dp.Out)
	}()
	return dp.Out
}

func (dp *DeforkersPool) defork(){
	for{
		select {
		case in1 := <- dp.In1:
			out, err := dp.In1Caster(in1)
			if err != nil{
				select {
					case dp.Out <- out:
					case <- dp.Done:
					return
				}
			}
			continue
		case in2 := <- dp.In2:
			out, err := dp.In2Caster(in2)
			if err != nil{
				select {
					case dp.Out <- out:
					case <- dp.Done:
					return
				}
			}
			continue
		}
	}
}
