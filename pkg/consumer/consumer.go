package consumer

import (
	"fmt"
	"sync"
	"time"

	"github.com/gargath/pleiades/pkg/sse"

	"github.com/gargath/pleiades/pkg/spinner"

	"github.com/gargath/pleiades/pkg/publisher/file"
)

var lastEventID string

// Consumer ingests an SSE stream from WMF and processes each event in turn
type Consumer struct {
	LastMsgID string
	stop      chan (bool)
	events    chan *sse.Event
	wg        sync.WaitGroup
	spinner   *spinner.Spinner
}

// Start begins consumption of the SSE stream
// If the current terminal is a TTY, it will output a progress spinner
func (c *Consumer) Start() (string, error) {
	c.stop = make(chan (bool))
	c.events = make(chan (*sse.Event))

	f, err := file.NewPublisher(c.events, "./events")
	if err != nil {
		return lastEventID, err
	}

	if !spinner.IsTTY() {
		fmt.Printf("Terminal is not a TTY, not displaying progress indicator")
	} else {
		c.spinner = spinner.NewSpinner("Processing... ")
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			for {
				select {
				case <-c.stop:
					return
				default:
					c.spinner.Tick()
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		eid, err := sse.Notify("https://stream.wikimedia.org/v2/stream/recentchange", c.events, c.stop)
		lastEventID = eid
		if err != nil && err == sse.ErrNilChan {
			fmt.Printf("Event consumer exited with error: %v", err)
		}
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		count, err := f.ReadAndPublish()
		if err != nil {
			fmt.Printf("File Publisher exited with error after processing %d events: %s", count, err)
		} else {
			fmt.Printf("File Publisher finished after processing %d events\n", count)
		}
	}()

	c.wg.Wait()
	return lastEventID, nil
}

// Stop will stop the consumer, close the connection and request all goroutines to exit
// It blocks until shutdown is complete
func (c *Consumer) Stop() {
	close(c.stop)
	close(c.events)
	c.wg.Wait()
}
