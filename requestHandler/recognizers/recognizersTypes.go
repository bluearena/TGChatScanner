package recognizers

import file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"

type PhotoRecognizersPool struct {
	In            chan *file.FileInfo
	Out           chan *RecognizedPhoto
	Done          chan struct{}
	WorkersNumber int
}

type RecognizedPhoto struct {
	Link  *file.FileLink
	Error error
}
