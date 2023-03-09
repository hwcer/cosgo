package xlsx

import (
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func NewDebouncer(after time.Duration) func(f func()) {
	d := &debouncer{after: after}
	return func(f func()) {
		d.add(f)
	}
}

type debouncer struct {
	mu    sync.Mutex
	after time.Duration
	timer *time.Timer
}

func (d *debouncer) add(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.after, f)
}

func recursiveWatch(w *fsnotify.Watcher, path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		err = w.Add(path)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
