// +build !darwin

package main

import (
	"log"

	"github.com/howeyc/fsnotify"
)

type fsnotifyWatcher struct {
	watcher *fsnotify.Watcher
	ch      chan *FileEvent
}

func NewWatcher(paths []string) (Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	ch := make(chan *FileEvent)
	for _, path := range paths {
		err = w.Watch(path)
		if err != nil {
			log.Fatalf("unable to watch '%s': %s", path, err)
		}
	}
	fw := &fsnotifyWatcher{
		watcher: w,
		ch:      ch,
	}
	go fw.handleEvents()
	return fw, nil
}

func (w *fsnotifyWatcher) Event() <-chan *FileEvent {
	return w.ch
}

func (w *fsnotifyWatcher) handleEvents() {
	for {
		select {
		case ev := <-w.watcher.Event:
			w.ch <- &FileEvent{Name: ev.Name}
		case err := <-w.watcher.Error:
			log.Println("error:", err)
		}
	}
}
