package main

import (
	"fmt"
	"os"
	"time"
)

// Event represents a single record
type Event struct {
	ID   int
	Data string
}

// EventIngestor handles buffered ingestion of events to a WAL file
type EventIngestor struct {
	file       *os.File
	filePath   string
	buffer     []Event
	bufferSize int
	flushTimer time.Duration
	ch         chan Event
}

// NewEventIngestor initializes the ingestor
func NewEventIngestor(filePath string, bufferSize int, flushTimer time.Duration) (*EventIngestor, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &EventIngestor{
		file:       f,
		filePath:   filePath,
		buffer:     make([]Event, 0, bufferSize),
		bufferSize: bufferSize,
		flushTimer: flushTimer,
		ch:         make(chan Event, bufferSize),
	}, nil
}

// Start begins ingestion: flushes buffer periodically and handles incoming events
func (ei *EventIngestor) Start() {
	ticker := time.NewTicker(ei.flushTimer)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case e := <-ei.ch:
				ei.buffer = append(ei.buffer, e)
				if len(ei.buffer) >= ei.bufferSize {
 					fmt.Println("batch_size reached")
					ei.flush()
				}
			case <-ticker.C:
				fmt.Println("ticker")
				if len(ei.buffer) > 0 {
					ei.flush()
				}
			}
		}
	}()
}

// Push adds a new event to the ingestor
func (ei *EventIngestor) Push(e Event) {
	ei.ch <- e
}

// flush writes buffered events to the file
func (ei *EventIngestor) flush() {
	for _, e := range ei.buffer {
		line := fmt.Sprintf("%d:%s\n", e.ID, e.Data)
		_, err := ei.file.WriteString(line)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
	}
	ei.file.Sync()
	ei.buffer = ei.buffer[:0]
}
