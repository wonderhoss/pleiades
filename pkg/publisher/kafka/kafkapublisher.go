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
)

const moduleName = "kafkapublisher"

var (
	logger      = log.MustGetLogger(moduleName)
	kafkaLogger = log.MustGetLogger("kafka-client")

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
	dest := viper.GetString("kafka.broker") // TODO: Remove viper references and use config struct instead
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

	f.w = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      f.destination.Brokers,
		Topic:        f.destination.Topic,
		BatchSize:    100,
		RequiredAcks: 0,
		Async:        true,
	})
	kc := PrometheusCollector{
		Publisher: f,
	}
	prometheus.DefaultRegisterer.MustRegister(kc)

	return f, nil
}

// ValidateConnection tests the connection to Kafka using the details given when creating the Publisher
func (f *Publisher) ValidateConnection() error {
	logger.Debug("Testing kafka connection")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialLeader(ctx, "tcp", f.destination.Brokers[0], f.destination.Topic, 0)
	if err != nil {
		return fmt.Errorf("Error connecting to leader for partition [0]: %v", err)
	}
	vs, err := conn.ApiVersions()
	if err != nil {
		return fmt.Errorf("Error retrieving api versions: %v", err)
	}
	logger.Debugf("Supported Kafka API versions")
	for _, ve := range vs {
		logger.Debugf("API version: %d; min: %d, max: %d", ve.ApiKey, ve.MinVersion, ve.MaxVersion)
	}
	return nil
}

// ReadAndPublish will read Events from the input channel and write them to the kafka topic
// configured for this Publisher.
//
// Calling ReadAndPublish() will reset the processed message counter of the underlying Publisher and
// returns the value of the counter when the Publisher's source channel is closed
func (f *Publisher) ReadAndPublish() (int64, error) {

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
	f.currMsgID = e.ID
	return nil
}

// GetResumeID will try to get the latest message published to Kafka and extract a resume ID from it
func (f *Publisher) GetResumeID() string {
	logger.Infof("Trying to retrieve resumable event ID from kafka")
	co1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel1()

	c, err := kafka.DialLeader(co1, "tcp", f.destination.Brokers[0], f.destination.Topic, 0)

	if err != nil {
		logger.Errorf("Error connecting to leader for partition [0]: %v", err)
		return ""
	}

	parts, err := c.ReadPartitions()
	if err != nil {
		logger.Errorf("Error reading partitions: %v", err)
	}
	for _, p := range parts {
		logger.Debugf("Found partition: %+v", p)
	}

	// TODO: dial leader for every partition and read latest offset from each, then compare messages to find truly freshest

	l2, err := c.ReadLastOffset()
	if err != nil {
		logger.Debugf("Error getting last offset: %v", err)
	}
	logger.Debugf("Last Offset found: %d", l2)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     f.destination.Brokers,
		Topic:       f.destination.Topic,
		ErrorLogger: &crudErrorLogger{},
		Logger:      newCrudLogger(),
	})

	r.SetOffset(l2 - 1)

	co, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	msg, err := r.ReadMessage(co)
	if err != nil {
		logger.Errorf("Error reading latest message: %v", err)
		return ""
	}
	key := string(msg.Key)
	return key
}
