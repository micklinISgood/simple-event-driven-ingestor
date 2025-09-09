package main

import (
	"fmt"
        "time"
)
func main() {
    walPath := "/data/wal.log"


    ingestor, err := NewEventIngestor(walPath, 10, 501*time.Millisecond)
    if err != nil {
	panic(err)
    }
    ingestor.Start()
    // Start file consumer
    fc, err := NewFileConsumer(walPath)
    if err != nil {
	panic(err)
    }
    fc.Start()
    go func() {
		for i := 1; i <= 50; i++ {
			ingestor.Push(Event{ID: i, Data: fmt.Sprintf("event-%d", i)})
			time.Sleep(50 * time.Millisecond)
		}
    }()
    	
    // Wait to ensure flushing happens
    time.Sleep(60 * time.Second)
    fmt.Println("Done writing events. Check /data/wal.log")
}
