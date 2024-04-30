package types

import (
	"io/fs"
	"time"

	"github.com/radovskyb/watcher"
)

type FileSystemChange struct {
	Op       watcher.Op
	FilePath string
	Fi       fs.FileInfo
}

type FileSystemRecord struct {
	FilePath string
	FileName string
	Mode     fs.FileMode
	Size     int64
}

type FileChangeLog struct {
	ChangeType  string    `json:"change_type"`
	FilePath    string    `json:"file_path"`
	Mode        string    `json:"mode"`
	Size        int64     `json:"size"`
	LastUpdated time.Time `json:"last_updated"`
}

type LogProgress struct {
	Type  string `json:"type"`
	Info  string `json:"info"`
	Error string `json:"error"`
}

type FChange uint32

const (
	Initial FChange = iota
	Add
	Update
	Delete
)

func (fc FChange) String() string {
	switch fc {
	case Initial:
		return "Initial"
	case Add:
		return "Add"
	case Update:
		return "Update"
	case Delete:
		return "Delete"
	default:
		return "Unknown"
	}
}
