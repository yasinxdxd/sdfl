package sdfl

import (
	"os"
	"time"
)

type FileWatcher struct {
	path          string
	lastWriteTime time.Time
}

func NewFileWatcher(path string) (*FileWatcher, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		path:          path,
		lastWriteTime: info.ModTime(),
	}, nil
}

func (fw *FileWatcher) HasChanged() (bool, error) {
	info, err := os.Stat(fw.path)
	if err != nil {
		return false, err
	}

	current := info.ModTime()
	if !current.Equal(fw.lastWriteTime) {
		fw.lastWriteTime = current
		return true, nil
	}
	return false, nil
}
