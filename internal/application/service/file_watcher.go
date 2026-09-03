package service

import (
	"log"
	"os"
	"time"
)

type FileWatcher struct {
	filePath  string
	onDelete  func()
	onCreate  func()
	lastExist bool
	stop      chan struct{}
}

func NewFileWatcher(filePath string, onDelete, onCreate func()) *FileWatcher {
	_, err := os.Stat(filePath)
	w := &FileWatcher{
		filePath:  filePath,
		onDelete:  onDelete,
		onCreate:  onCreate,
		lastExist: err == nil,
		stop:      make(chan struct{}),
	}
	go w.run()
	return w
}

func (w *FileWatcher) Stop() {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
}

func (w *FileWatcher) run() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
		}
		_, err := os.Stat(w.filePath)
		exists := err == nil

		if exists && !w.lastExist {
			log.Printf("file created: %s", w.filePath)
			if w.onCreate != nil {
				w.onCreate()
			}
		} else if !exists && w.lastExist {
			log.Printf("file deleted: %s", w.filePath)
			if w.onDelete != nil {
				w.onDelete()
			}
		}
		w.lastExist = exists
	}
}
