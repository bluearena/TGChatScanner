package poolTypes

import (
	file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"
	"github.com/zwirec/TGChatScanner/requestHandler/recognizers"
)

type FromDownloadedToComplete func(*file.DownloadedFile) (*file.CompleteFile, error)

type FromRecognizedToComplete func(*recognizers.RecognizedPhoto) (*file.CompleteFile, error)
