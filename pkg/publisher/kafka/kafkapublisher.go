package kafka

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	kafka "github.com/segmentio/kafka-go"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/publisher"
	"github.com/gargath/pleiades/pkg/sse"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const moduleName = "kafkapublisher"

// TODO: reuse these with additional publisher label rather than having separate counters
var (
	eventsPublished = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pleiades_kafka_publish_events_total",
			Help: "The total number of events published to kafka",
		})

	pubErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pleiades_kafka_publish_errors_total",
			Help: "Total numbers of errors encountered while publishing to kafka",
		},
		[]string{"type"})

	kafkaWriteTime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pleiades_kafka_writer_write_time",
			Help: "Time the kafka writer spent writing",
		},
		[]string{"agg"})

	kafkaWaitTime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pleiades_kafka_writer_wait_time",
			Help: "Time the kafka writer spent waiting",
		},
		[]string{"agg"})

	logger = log.MustGetLogger(moduleName)
)

func init() {
	flag.Bool("kafka.enable", false, "enable the kafka publisher")
	flag.String("kafka.broker", "localhost:9092", "the kafka broker to connect to")
	flag.String("kafka.topic", "pleiades-events", "the kafka topic to publish to")
}

// NewPublisher returns a Publisher initialized with the source channel and kafka destination provided
func NewPublisher(src <-chan *sse.Event) (publisher.Publisher, error) {
	dest := viper.GetString("kafka.broker")
	topic := viper.GetString("kafka.topic")
	if src == nil {
		return nil, ErrNilChan
	}
	o := &ConnectionOpts{
		Brokers: []string{dest},
		Topic:   topic,
	}
	f := &Publisher{
		source:      src,
		destination: o,
	}
	return f, nil
}

// ReadAndPublish will read Events from the input channel and write them to the kafka topic
// configured for this Publisher.
//
// Calling ReadAndPublish() will reset the processed message counter of the underlying Publisher and
// returns the value of the counter when the Publisher's source channel is closed
func (f *Publisher) ReadAndPublish() (int64, error) {
	f.w = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      f.destination.Brokers,
		Topic:        f.destination.Topic,
		BatchSize:    100,
		RequiredAcks: 0,
		Async:        true,
	})
	// TODO: Use Prometheus Collector instead of this jank
	done := make(chan (bool))
	go func() {
		logger.Debug("Starting Kafka Prometheus Exporter")
		for {
			select {
			case <-done:
				logger.Debug("Stopping Kafka Prometheus Exporter")
				return
			default:
				time.Sleep(5 * time.Second)
				stats := f.w.Stats()
				kafkaWriteTime.WithLabelValues("min").Set(stats.WriteTime.Min.Seconds())
				kafkaWriteTime.WithLabelValues("max").Set(stats.WriteTime.Max.Seconds())
				kafkaWriteTime.WithLabelValues("avg").Set(stats.WriteTime.Avg.Seconds())
				kafkaWaitTime.WithLabelValues("min").Set(stats.WaitTime.Min.Seconds())
				kafkaWaitTime.WithLabelValues("max").Set(stats.WaitTime.Max.Seconds())
				kafkaWaitTime.WithLabelValues("avg").Set(stats.WaitTime.Avg.Seconds())
			}
		}
	}()
	logger.Debug("Kafka publisher starting to process events")
	f.msgCount = 0
	for e := range f.source {
		f.msgCount++
		if e != nil {
			err := f.ProcessEvent(e)
			if err != nil {
				return f.msgCount, fmt.Errorf("error processing event: %v", err)
			}
		}
	}
	close(done)
	logger.Debug("Kafka publisher stopped")
	return f.msgCount, nil
}

// ProcessEvent writes a single event to a kafka
func (f *Publisher) ProcessEvent(e *sse.Event) error {
	eventsPublished.Inc()
	d, err := ioutil.ReadAll(e.GetData())
	if err != nil {
		pubErrors.WithLabelValues("event_data_read").Inc()
		return fmt.Errorf("error reading event data: %v", err)
	}
	err = f.w.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(e.ID),
		Value: d,
	})
	if err != nil {
		pubErrors.WithLabelValues("write").Inc()
		return fmt.Errorf("error writing to kafka: %v", err)
	}
	return nil
}
