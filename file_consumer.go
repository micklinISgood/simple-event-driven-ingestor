package main

import (
	"bufio"
	"fmt"
	"os"
        "time"
	"sync"
	"github.com/fsnotify/fsnotify"
)
type FileConsumer struct {
	filePath string
	offset   int64
	file     *os.File
	reader   *bufio.Reader
	watcher  *fsnotify.Watcher
	done     chan struct{}
	ingestorDone <-chan struct{}
	mu 	sync.Mutex
}

func NewFileConsumer(filePath string, ingestorDone <-chan struct{}) (*FileConsumer, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		f.Close()
		return nil, err
	}

	err = watcher.Add(filePath)
	if err != nil {
		f.Close()
		watcher.Close()
		return nil, err
	}

	return &FileConsumer{
		filePath: filePath,
		file:     f,
		reader:   bufio.NewReader(f),
		watcher:  watcher,
		done:     make(chan struct{}),
		offset:   0,
		ingestorDone: ingestorDone,
	}, nil
}


func (fc *FileConsumer) consumeNewLines(tag string) {
    fc.mu.Lock()
    defer fc.mu.Unlock()
    // Move to the current offset
    fc.file.Seek(fc.offset, os.SEEK_SET)
    fc.reader.Reset(fc.file)

    for {
        line, err := fc.reader.ReadString('\n')
	//fmt.Println(err)
        if err != nil {
            break
        }
        fc.offset += int64(len(line))
        fmt.Printf("Consumer (%s): %s is ready to produce to kafka.\n", tag, line)
    }
}



func (fc *FileConsumer) Start() {
	defer fc.file.Close()
	defer fc.watcher.Close()

	// Read existing lines
	fc.file.Seek(0, os.SEEK_SET)
	for {
		line, err := fc.reader.ReadString('\n')
		if err != nil {
			break
		}
		fc.offset += int64(len(line))
		fmt.Println("Consumer (existing):", line)
	}

	// Tail new writes
	for {
		select {
		case event := <-fc.watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write{
				fc.consumeNewLines("watcher") 
			}
		case err := <-fc.watcher.Errors:
			fmt.Println("Watcher error:", err)
		case <-fc.ingestorDone:
			time.Sleep(1 * time.Second)
			fc.consumeNewLines("stop")
			return
		}
	}
}


func (fc *FileConsumer) Stop() {
	close(fc.done)
}
