package kafka

import (
	"fmt"

	kafka "github.com/segmentio/kafka-go"

	"github.com/gargath/pleiades/pkg/ingester/sse"
)

// Publisher reads Events and writes them to disk
type Publisher struct {
	destination *ConnectionOpts
	source      <-chan *sse.Event
	msgCount    int64
	w           *kafka.Writer
	currMsgID   string
}

// Opts hold configuration for the kafka publisheru
type Opts struct {
	Broker string
	Topic  string
}

// ConnectionOpts wrap the information needed to connect to kafka
type ConnectionOpts struct {
	Brokers []string
	Topic   string
}

// ErrNilChan indicates that the FilePublisher has no source channel
var ErrNilChan error = fmt.Errorf("Source channel is nil")
