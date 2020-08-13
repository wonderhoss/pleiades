package ingester

import (
	"github.com/gargath/pleiades/pkg/ingester/publisher/file"
	"github.com/gargath/pleiades/pkg/ingester/publisher/kafka"
	"github.com/gargath/pleiades/pkg/ingester/sse"
	"github.com/gargath/pleiades/pkg/util"
)

// Coordinator ingests an SSE stream from WMF and processes each event in turn
type Coordinator struct {
	LastMsgID string
	Resume    bool
	File      *file.Opts
	Kafka     *kafka.Opts
	stop      chan (bool)
	events    chan *sse.Event
	spinner   *util.Spinner
}
