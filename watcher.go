package main

type Watcher interface {
	Event() <-chan *FileEvent
}

type FileEvent struct {
	Name string
}
