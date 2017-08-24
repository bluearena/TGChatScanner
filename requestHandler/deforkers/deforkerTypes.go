package deforkers

import (
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"github.com/zwirec/TGChatScanner/requestHandler/poolTypes"
	"github.com/zwirec/TGChatScanner/requestHandler/recognizers"
)

type DeforkersPool struct {
	In1              chan *file.DownloadedFile
	In2              chan *recognizers.RecognizedPhoto
	Out              chan *file.CompleteFile
	DeforkDownloaded poolTypes.FromDownloadedToComplete
	DeforkRecognized poolTypes.FromRecognizedToComplete
	Done             chan struct{}
	WorkersNumber    int
}
