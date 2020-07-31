package consumer

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	//	sse "astuart.co/go-sse"
	sse "github.com/gargath/pleiades/pkg/sse"

	"github.com/gargath/pleiades/pkg/spinner"
)

// Consumer ingests an SSE stream from WMF and processes each event in turn
type Consumer struct {
	MsgReceived int
	MsgRead     int
	LastMsgID   string
	stop        chan (bool)
	events      chan *sse.Event
	wg          sync.WaitGroup
	spinner     *spinner.Spinner
}

// Start begins consumption of the SSE stream
// If the current terminal is a TTY, it will output a progress spinner
func (c *Consumer) Start() {
	c.stop = make(chan (bool))
	c.events = make(chan (*sse.Event))
	if !spinner.IsTTY() {
		fmt.Printf("Terminal is not a TTY, not displaying progress indicator")
	} else {
		c.spinner = spinner.NewSpinner("Working... ")
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			for {
				select {
				case <-c.stop:
					return
				default:
					c.spinner.Tick(fmt.Sprintf("Received: %d, Read: %d", c.MsgReceived, c.MsgRead))
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		err := sse.Notify("https://stream.wikimedia.org/v2/stream/recentchange", c.events, c.stop)
		if err != nil && err == sse.ErrNilChan {
			panic(err)
		}
	}()
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		o, err := os.Stat("events")
		if os.IsNotExist(err) {
			errDir := os.MkdirAll("events", 0755)
			if errDir != nil {
				panic(err)
			}
		} else if o.Mode().IsRegular() {
			panic(fmt.Errorf("events directory exists as file"))
		}
		for {
			select {
			case e := <-c.events:
				c.MsgReceived++
				if e != nil {
					d, err := ioutil.ReadAll(e.GetData())
					if err != nil {
						fmt.Printf("Error reading msg: %v\n", err)
					}
					err = ioutil.WriteFile(fmt.Sprintf("events/event-%d.dat", c.MsgRead), d, 0644)
					if err != nil {
						fmt.Printf("Error writing msg to file: %v\n", err)
					}
					c.MsgRead++
					c.LastMsgID = e.ID
				}
			case <-c.stop:
				fmt.Printf("Last message consumed: %s\n", c.LastMsgID)
				return
			}
		}
	}()
	c.wg.Wait()
}

// Stop will stop the consumer, close the connection and request all goroutines to exit
// It blocks until shutdown is complete
func (c *Consumer) Stop() {
	close(c.stop)
	close(c.events)
	c.wg.Wait()
}
