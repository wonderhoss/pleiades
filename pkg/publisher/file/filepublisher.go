package file

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gargath/pleiades/pkg/publisher"
	"github.com/gargath/pleiades/pkg/sse"
)

// NewPublisher returns a Publisher initialized with the source channel and destination path provided
func NewPublisher(src <-chan *sse.Event, dest string) (publisher.Publisher, error) {
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
			panic(err)
		}
	} else if o.Mode().IsRegular() {
		panic(fmt.Errorf("destination path %s exists as file", dest))
	}
	f := &Publisher{
		source:      src,
		destination: dest,
	}
	return f, nil
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
	return f.msgCount, nil
}

// ProcessEvent writes a single event to a file
func (f *Publisher) ProcessEvent(e *sse.Event) error {
	d, err := ioutil.ReadAll(e.GetData())
	if err != nil {
		return fmt.Errorf("error reading event data: %v", err)
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/event-%d.dat", f.destination, f.msgCount), d, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}
	return nil
}
