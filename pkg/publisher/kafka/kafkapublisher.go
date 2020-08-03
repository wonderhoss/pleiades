package kafka

import (
	"context"
	"fmt"
	"io/ioutil"

	kafka "github.com/segmentio/kafka-go"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/publisher"
	"github.com/gargath/pleiades/pkg/sse"
	"github.com/prometheus/client_golang/prometheus"
)

const moduleName = "kafkapublisher"

var (
	logger = log.MustGetLogger(moduleName)

	pubErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pleiades_kafka_writer_errors_total",
		Help: "Total numbers of errors encountered while publishing to kafka",
	},
		[]string{"type"},
	)
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
	kc := PrometheusCollector{
		Writer: f.w,
	}
	prometheus.DefaultRegisterer.MustRegister(kc)

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
	logger.Debug("Kafka publisher stopped")
	return f.msgCount, nil
}

// ProcessEvent writes a single event to a kafka
func (f *Publisher) ProcessEvent(e *sse.Event) error {
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
