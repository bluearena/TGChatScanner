package filetypes

import (
	"context"
	"time"
)

const (
	Undefiend       int32 = 0
	RecSuccess      int32 = 1
	RecFailed       int32 = 2
	DownloadSuccess int32 = 4
	DownloadFailed  int32 = 8
)

type FileBasic struct {
	FileId       string
	Type         string
	Sent         time.Time
	From         int64
	Tags         []string
	Errorc       chan error
	BasicContext context.Context
}

type FileLink struct {
	FileDownloadURL string
	LocalPath       string
	Basics          *FileBasic
	Status          *int32
}

type PreparedFile struct {
	Link  *FileLink
	Error error
}

type DownloadedFile struct {
	Link  *FileLink
	Error error
}

type CompleteFile FileLink

type FileInfo FileLink
