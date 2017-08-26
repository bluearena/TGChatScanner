package forkers

import file "github.com/zwirec/TGChatScanner/requestHandler/filetypes"

type ForkersPool struct {
	In             chan *file.PreparedFile
	Out1           chan *file.FileLink
	Out2           chan *file.FileInfo
	Done           chan struct{}
	ForkToFileLink InToFileLink
	ForkToFileInfo InToFileInfo
	WorkersNumber  int
}

type InToFileLink func(*file.PreparedFile) (*file.FileLink, error)

type InToFileInfo func(result *file.PreparedFile) (*file.FileInfo, error)
