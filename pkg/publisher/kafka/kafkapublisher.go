package kafka

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"time"

	kafka "github.com/segmentio/kafka-go"

	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/publisher"
	"github.com/gargath/pleiades/pkg/sse"
	"github.com/prometheus/client_golang/prometheus"
)

const moduleName = "kafkapublisher"

var (
	logger      = log.MustGetLogger(moduleName)
	kafkaLogger = log.MustGetLogger("kafka-client")

	timeStampRegExp = regexp.MustCompile(`"timestamp":([0-9]+).*`)

	pubErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pleiades_kafka_writer_errors_total",
		Help: "Total numbers of errors encountered while publishing to kafka",
	},
		[]string{"type"},
	)
)

// NewPublisher returns a Publisher initialized with the source channel and kafka destination provided
func NewPublisher(opts *Opts, src <-chan *sse.Event) (publisher.Publisher, error) {
	dest := opts.Broker
	topic := opts.Topic
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
		Balancer:     kafka.Murmur2Balancer{},
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

	// TODO: Dial all brokers
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = f.w.WriteMessages(ctx, kafka.Message{
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
	co1, cancel1 := context.WithTimeout(context.Background(), 30*time.Second)
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
	logger.Debugf("Read list of partitions from kafka")
	latest, err := f.findLatestMessage(parts)
	if err != nil {
		logger.Infof("Error fetching Resume ID: %v", err)
		return ""
	}
	return string(latest.Key)
}

func (f *Publisher) findLatestMessage(partitions []kafka.Partition) (*kafka.Message, error) {
	for _, p := range partitions {
		logger.Debugf("Scanning t partition for latest messages: %+v", p)
	}

	// Ask each partition in parallel for latest message and collect
	messages := make([]*kafka.Message, len(partitions))
	messageErrors := []error{}
	msgChan := make(chan (*kafka.Message))
	errChan := make(chan (error))
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	for _, p := range partitions {
		go f.getLatestMessageForPartition(ctx, p, msgChan, errChan)
	}
	for i := 0; i < len(partitions); i++ {
		select {
		case m := <-msgChan:
			messages[i] = m
		case e := <-errChan:
			messageErrors = append(messageErrors, e)
		}
	}
	if len(messageErrors) > 0 {
		logger.Errorf("Errors during partition scan:")
		for _, e := range messageErrors {
			logger.Errorf(" - %v", e)
		}
		return nil, fmt.Errorf("unable to retrieve latest offset due to errors encountered during partition scan")
	}

	// Iterate over latest messages from each partition to find most recent timestamp
	var latest *kafka.Message
	var latestTS int64 = 0
	for _, m := range messages {
		if latest == nil {
			latest = m
		} else {
			match := timeStampRegExp.FindStringSubmatch(string(m.Key))
			if len(match) < 2 {
				logger.Errorf("Event ID %s has no timestamp", m.Key)
				continue
			}
			timeStamp, err := strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				logger.Errorf("Error parsing timestamp %s: %v", match[1], err)
				continue
			}
			if timeStamp > latestTS {
				latest = m
			}
		}
	}
	return latest, nil
}

func (f *Publisher) getLatestMessageForPartition(ctx context.Context, p kafka.Partition, m chan<- (*kafka.Message), e chan<- (error)) {
	c, err := kafka.DialLeader(ctx, "tcp", f.destination.Brokers[0], f.destination.Topic, p.ID)
	if err != nil {
		e <- fmt.Errorf("Error connecting to leader for partition %d: %v", p.ID, err)
		return
	}
	l, err := c.ReadLastOffset()
	if err != nil {
		e <- fmt.Errorf("Error getting last offset from partition %d: %v", p.ID, err)
		return
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     f.destination.Brokers,
		Topic:       f.destination.Topic,
		ErrorLogger: &crudErrorLogger{},
		Logger:      newCrudLogger(),
		Partition:   p.ID,
	})
	r.SetOffset(l - 1)
	msg, err := r.ReadMessage(ctx)
	if err != nil {
		e <- fmt.Errorf("Error reading latest message from partition %d: %v", p.ID, err)
		return
	}
	r.Close()
	m <- &msg
}
