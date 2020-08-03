package kafka

import (
	"fmt"

	kafka "github.com/segmentio/kafka-go"

	"github.com/gargath/pleiades/pkg/sse"
)

// Publisher reads Events and writes them to disk
type Publisher struct {
	destination *ConnectionOpts
	source      <-chan *sse.Event
	msgCount    int64
	w           *kafka.Writer
}

// ConnectionOpts wrap the information needed to connect to kafka
type ConnectionOpts struct {
	Brokers []string
	Topic   string
}

// ErrNilChan indicates that the FilePublisher has no source channel
var ErrNilChan error = fmt.Errorf("Source channel is nil")
