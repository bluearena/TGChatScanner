package filetypes

import "time"

type FileBasic struct {
	FileId string
	Type   string
	Sent   time.Time

	Context map[string]interface{}
}

type FileLink struct {
	FileDownloadURL string
	LocalPath       string
	Basics          *FileBasic
}
type PreparedFile struct {
	Link  *FileLink
	Error error
}

type DownloadedFile PreparedFile

type CompleteFile FileLink

type FileInfo struct {
	FileId  string
	FileURL string
}
