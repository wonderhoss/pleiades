// This package includes code Copyright (c) 2015 Andrew Stuart,
// licensed under MIT

package sse

import "fmt"

//SSE name constants
const (
	eName = "event"
	dName = "data"
	iName = "id"
)

var (
	//ErrNilChan will be returned by Notify if it is passed a nil channel
	ErrNilChan = fmt.Errorf("nil channel given")

	delim = []byte{':', ' '}
)
