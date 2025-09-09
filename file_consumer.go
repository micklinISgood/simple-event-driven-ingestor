package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

type FileConsumer struct {
	filePath string
	offset   int64
	file     *os.File
	reader   *bufio.Reader
	watcher  *fsnotify.Watcher
}

// NewFileConsumer initializes the consumer
func NewFileConsumer(filePath string) (*FileConsumer, error) {
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
		offset:   0,
	}, nil
}

// Start begins consuming the file: first existing lines, then tail new writes
func (fc *FileConsumer) Start() {
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

	// Start tailing new writes
	go func() {
		defer fc.file.Close()
		defer fc.watcher.Close()

		for {
			select {
			case event := <-fc.watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					fc.file.Seek(fc.offset, os.SEEK_SET)
					fc.reader.Reset(fc.file)
					for {
						line, err := fc.reader.ReadString('\n')
						if err != nil {
							break
						}
						fc.offset += int64(len(line))
						fmt.Println("Consumer (new):", line)
					}
				}
			case err := <-fc.watcher.Errors:
				fmt.Println("Watcher error:", err)
			}
		}
	}()
}
