package coordinator

import (
	"github.com/gargath/pleiades/pkg/publisher/file"
	"github.com/gargath/pleiades/pkg/publisher/kafka"
	"github.com/gargath/pleiades/pkg/spinner"
	"github.com/gargath/pleiades/pkg/sse"
)

// Coordinator ingests an SSE stream from WMF and processes each event in turn
type Coordinator struct {
	LastMsgID string
	Resume    bool
	File      *file.Opts
	Kafka     *kafka.Opts
	stop      chan (bool)
	events    chan *sse.Event
	spinner   *spinner.Spinner
}
