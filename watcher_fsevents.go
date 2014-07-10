// +build darwin

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ronbu/fsevents"
)

type fseventsWatcher struct {
	stream  *fsevents.Stream
	ch      chan *FileEvent
	pathSet map[string]bool
}

func NewWatcher(paths []string) (Watcher, error) {
	pathSet := map[string]bool{}
	var root string
	for _, p := range paths {
		p, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		if root == "" {
			root = p
		} else {
			root = commonDir(root, p)
		}
		pathSet[p[1:]] = true
	}

	// TODO what if there are multiple devices? Need root per device
	fi, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	dev := fsevents.Device(fi.Sys().(*syscall.Stat_t).Dev)

	stream := fsevents.New(dev, fsevents.NOW, 1*time.Second, fsevents.CF_FILEEVENTS, root)
	ch := make(chan *FileEvent)
	fw := &fseventsWatcher{
		stream:  stream,
		ch:      ch,
		pathSet: pathSet,
	}
	go fw.handleEvents()
	if !stream.Start() {
		return nil, fmt.Errorf("Error starting fsevents.Stream")
	}
	return fw, nil
}

func (w *fseventsWatcher) Event() <-chan *FileEvent {
	return w.ch
}

func (w *fseventsWatcher) handleEvents() {
	for {
		select {
		case evs := <-w.stream.Chan:
			for _, ev := range evs {
				if ev.Flags&fsevents.EF_ISFILE != 0 {
					p := ev.Path
					if w.pathSet[p] || w.pathSet[filepath.Dir(p)] {
						w.ch <- &FileEvent{Name: p}
					}
				}
			}
		}
	}
}

// Find the common root directory of two paths
func commonDir(a, b string) string {
	a += "/"
	b += "/"

	m := a
	if len(a) > len(b) {
		m = b
	}
	x := 0
	for i := 0; i < len(m); i++ {
		if a[i] != b[i] {
			break
		}
		if a[i] == '/' {
			x = i
		}
	}
	return m[:x]
}
