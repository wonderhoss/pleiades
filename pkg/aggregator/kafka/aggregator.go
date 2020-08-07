package kafka

import (
	"sync"
	"time"

	"github.com/gargath/pleiades/pkg/log"
)

const moduleName = "kafka-agg"

var (
	wg sync.WaitGroup

	logger = log.MustGetLogger(moduleName)
)

// Start starts up the aggregation server
func (a *Aggregator) Start() error {
	a.stop = make(chan (bool))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-a.stop:
				{
					return
				}
			default:
				err := a.run()
				if err != nil {
					logger.Errorf("Aggregator exited with error: %v", err)
				}
			}
		}
	}()

	wg.Wait()
	return nil
}

// Stop shuts down the aggregation server
func (a *Aggregator) Stop() {
	close(a.stop)
	wg.Wait()
}

func (a *Aggregator) run() error {
	for {
		select {
		case <-a.stop:
			return nil
		default:
			logger.Info("...runtick...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
