package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	walPath := "/data/wal.log"

	var wg sync.WaitGroup

	// Initialize EventIngestor
	ingestor, err := NewEventIngestor(walPath, 10, 501*time.Millisecond)
	if err != nil {
		panic(err)
	}
	ingestor.Start() // internally launches flushing goroutine

	// Initialize FileConsumer
	fc, err := NewFileConsumer(walPath)
	if err != nil {
		panic(err)
	}

	// Add 2 to WaitGroup: consumer + producer
	wg.Add(2)

	// Start consumer
	go func() {
		defer wg.Done()
		fc.Start()
	}()

	// Start producer
	go func() {
		defer wg.Done()
		for i := 1; i <= 50; i++ {
			ingestor.Push(Event{ID: i, Data: fmt.Sprintf("event-%d", i)})
			time.Sleep(50 * time.Millisecond)
		}
		fc.Stop()
	}()

	// Wait for both goroutines to finish
	wg.Wait()
	fmt.Println("All events produced and consumed. Check /data/wal.log")
}
