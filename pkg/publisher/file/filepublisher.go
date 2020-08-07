package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/publisher"
	"github.com/gargath/pleiades/pkg/sse"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const moduleName = "filepublisher"

var (
	eventsPublished = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pleiades_file_publish_events_total",
			Help: "The total number of events published to filesystem",
		})

	pubErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pleiades_file_publish_errors_total",
			Help: "Total numbers of errors encountered while publishing to filesystem",
		},
		[]string{"type"})

	logger = log.MustGetLogger(moduleName)
)

// NewPublisher returns a Publisher initialized with the source channel and destination path provided
func NewPublisher(opts *Opts, src <-chan *sse.Event) (publisher.Publisher, error) {
	dest := opts.Destination
	if src == nil {
		return nil, ErrNilChan
	}
	if dest == "" {
		return nil, ErrNoDest
	}
	o, err := os.Stat(dest)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dest, 0755)
		if errDir != nil {
			logger.Fatalf("failed to create destination directory: %v", errDir)
			panic(err)
		}
	} else if o.Mode().IsRegular() {
		logger.Errorf("destination path %s exists and is file", dest)
		return nil, fmt.Errorf("destination path %s exists as file", dest)
	}
	uid := strconv.FormatInt(time.Now().Unix(), 10)
	f := &Publisher{
		source:      src,
		destination: dest,
		prefix:      uid,
	}
	return f, nil
}

// ValidateConnection always returns nil and only serves to satisfy the Publisher interface
func (f *Publisher) ValidateConnection() error {
	return nil
}

// ReadAndPublish will read Events from the input channel and write them to file
// File names are sequential and relative to the destination directory
// If the FilePublisher's destionation directory is not set, ReadAndPublish returns ErrNoDest
//
// Calling ReadAndPublish() will reset the processed message counter of the underlying Publisher and
// returns the value of the counter when the Publisher's source channel is closed
func (f *Publisher) ReadAndPublish() (int64, error) {
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
	err := ioutil.WriteFile("./.pleiades_resumeID", []byte(f.lastEventID), 0644)
	if err != nil {
		logger.Error("unable to write last processed event ID to file .pleiades_resumeID: %v", err)
	}
	return f.msgCount, nil
}

// ProcessEvent writes a single event to a file
func (f *Publisher) ProcessEvent(e *sse.Event) error {
	eventsPublished.Inc()
	d, err := ioutil.ReadAll(e.GetData())
	if err != nil {
		pubErrors.WithLabelValues("event_data_read").Inc()
		return fmt.Errorf("error reading event data: %v", err)
	}
	d = append([]byte("\n"), d...)
	d = append([]byte(e.ID), d...)
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s-event-%d.dat", f.destination, f.prefix, f.msgCount), d, 0644)
	if err != nil {
		pubErrors.WithLabelValues("write").Inc()
		return fmt.Errorf("error writing file: %v", err)
	}
	f.lastEventID = e.ID
	return nil
}

// GetResumeID attempts to read the ID of the last processed event from disk and returns it
func (f *Publisher) GetResumeID() string {

	data, err := ioutil.ReadFile("./.pleiades_resumeID")

	if err != nil {
		logger.Errorf("failed to open resume ID file .pleiades_resumeID: %v", err)
		return ""
	}
	return string(data)
}
