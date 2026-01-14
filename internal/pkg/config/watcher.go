package config

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// fileWatcher encapsulates file watching logic with debouncing
type fileWatcher struct {
	w           *fsnotify.Watcher
	files       []string
	reloadFunc  func() error
	onError     func(error)
	debounceDur time.Duration
}

// newFileWatcher creates a new file watcher for the given files
func newFileWatcher(files []string, reloadFunc func() error, onError func(error)) (*fileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	fw := &fileWatcher{
		w:           w,
		files:       files,
		reloadFunc:  reloadFunc,
		onError:     onError,
		debounceDur: 100 * time.Millisecond,
	}

	// Watch all unique directories containing config files
	watchedDirs := make(map[string]bool)
	for _, f := range files {
		absPath, err := filepath.Abs(f)
		if err != nil {
			continue
		}
		dir := filepath.Dir(absPath)
		if watchedDirs[dir] {
			continue
		}
		if err := w.Add(dir); err != nil {
			if onError != nil {
				onError(fmt.Errorf("failed to watch directory %s: %w", toRelativePath(dir), err))
			}
			continue
		}
		watchedDirs[dir] = true
	}

	go fw.watchLoop()
	return fw, nil
}

// Close closes the underlying watcher
func (fw *fileWatcher) Close() error {
	return fw.w.Close()
}

// watchLoop monitors file events and triggers reloads with debouncing
func (fw *fileWatcher) watchLoop() {
	var timer *time.Timer

	for {
		select {
		case event, ok := <-fw.w.Events:
			if !ok {
				return
			}
			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) && !event.Has(fsnotify.Rename) {
				continue
			}

			// Check if event file matches any of our config files
			matches := false
			absEventPath, _ := filepath.Abs(event.Name)
			for _, f := range fw.files {
				absConfigPath, _ := filepath.Abs(f)
				if absEventPath == absConfigPath {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}

			// Debounce and reload
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(fw.debounceDur, func() {
				if err := fw.reloadFunc(); err != nil && fw.onError != nil {
					fw.onError(err)
				}
			})

		case err, ok := <-fw.w.Errors:
			if !ok {
				return
			}
			if fw.onError != nil {
				fw.onError(err)
			}
		}
	}
}
