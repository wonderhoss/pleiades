package consumer

import (
	"fmt"
	"sync"
	"time"

	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/publisher/file"
	"github.com/gargath/pleiades/pkg/publisher/kafka"
	"github.com/gargath/pleiades/pkg/spinner"
	"github.com/gargath/pleiades/pkg/sse"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
)

const moduleName = "consumer"

var (
	lastEventID string
	wgPub       sync.WaitGroup
	wgSub       sync.WaitGroup

	restarts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pleiades_goroutine_restarts",
			Help: "Total numbers of restarts of component goroutines",
		},
		[]string{"component"})

	logger = log.MustGetLogger(moduleName)
)

// Consumer ingests an SSE stream from WMF and processes each event in turn
type Consumer struct {
	LastMsgID string
	stop      chan (bool)
	events    chan *sse.Event
	spinner   *spinner.Spinner
}

// Start begins consumption of the SSE stream
// If the current terminal is a TTY, it will output a progress spinner
func (c *Consumer) Start() (string, error) {
	c.stop = make(chan (bool))
	c.events = make(chan (*sse.Event))

	if !spinner.IsTTY() {
		logger.Info("Terminal is not a TTY, not displaying progress indicator")
	} else {
		c.spinner = spinner.NewSpinner("Processing... ")
		wgPub.Add(1)
		go func() {
			defer wgPub.Done()
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

	wgPub.Add(1)
	go func() {
		defer wgPub.Done()
		for {
			select {
			case <-c.stop:
				{
					return
				}
			default:
				{
					eid, err := sse.Notify("https://stream.wikimedia.org/v2/stream/recentchange", c.events, c.stop)
					restarts.WithLabelValues("wmf_consumer").Inc()
					lastEventID = eid
					if err != nil {
						logger.Errorf("Event consumer exited with error: %v", err)
					}
				}
			}
		}
	}()

	if viper.GetBool("file.enable") {
		f, err := file.NewPublisher(c.events)
		if err != nil {
			return lastEventID, fmt.Errorf("Failed to initialize file publisher: %v", err)
		}
		wgSub.Add(1)
		go func() {
			defer wgSub.Done()
			for {
				for {
					select {
					case <-c.stop:
						{
							return
						}
					default:
						count, err := f.ReadAndPublish()
						logger.Debug("File Publisher exited")
						if err != nil {
							logger.Errorf("File Publisher exited with error after processing %d events: %s", count, err)
						} else {
							logger.Infof("File Publisher finished after processing %d events\n", count)
						}
						restarts.WithLabelValues("file_publisher").Inc()
					}
				}
			}
		}()
	}

	if viper.GetBool("kafka.enable") {
		k, err := kafka.NewPublisher(c.events)
		if err != nil {
			return lastEventID, fmt.Errorf("Failed to initialize kafka publisher: %v", err)
		}
		wgSub.Add(1)
		go func() {
			defer wgSub.Done()
			for {
				for {
					select {
					case <-c.stop:
						{
							return
						}
					default:
						count, err := k.ReadAndPublish()
						logger.Debug("Kafka Publisher exited")
						if err != nil {
							logger.Errorf("Kafka Publisher exited with error after processing %d events: %s", count, err)
						} else {
							logger.Infof("Kafka Publisher finished after processing %d events\n", count)
						}
						restarts.WithLabelValues("kafka_publisher").Inc()
					}
				}
			}
		}()
	}

	wgSub.Wait()
	return lastEventID, nil
}

// Stop will stop the consumer, close the connection and request all goroutines to exit
// It blocks until shutdown is complete
func (c *Consumer) Stop() {
	close(c.stop)
	wgPub.Wait()
	close(c.events)
	wgSub.Wait()
}
