// This package includes code Copyright (c) 2015 Andrew Stuart,
// licensed under MIT

package sse

import (
	"bytes"
	"io"
)

// Event is a go representation of an http server-sent event
type Event struct {
	URI  string
	Type string
	ID   string //me
	data *bytes.Buffer
}

// GetData returns a read-only view of this Event's data buffer
func (e *Event) GetData() io.Reader {
	return e.data
}
