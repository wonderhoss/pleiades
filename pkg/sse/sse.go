// This package includes code Copyright (c) 2015 Andrew Stuart,
// licensed under MIT

package sse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

var client = &http.Client{}

func liveReq(verb, uri string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(verb, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/event-stream")

	return req, nil
}

func parseLine(bs []byte, currEvent *Event) {
	spl := bytes.SplitN(bs, delim, 2)
	if len(spl) < 2 {
		if spl[0][0] == 0x003A { // a colon (:) - means this is a comment in the stream
			return
		}
		// else: log non-compliant line and continue
	}
	switch string(spl[0]) {
	case iName:
		currEvent.ID = string(bytes.TrimSpace(spl[1]))
	case eName:
		currEvent.Type = string(bytes.TrimSpace(spl[1]))
	case dName:
		if currEvent.data.Len() > 0 {
			currEvent.data.WriteByte(byte(0x000A))
		}
		currEvent.data.Write(bytes.TrimSpace(spl[1]))
	}
}

//Notify takes the uri of an SSE stream and channel, and will send an Event
//down the channel when recieved, until the stream is closed. It will then
//close the stream. This is blocking, and so you will likely want to call this
//in a new goroutine (via `go Notify(..)`)
func Notify(uri string, evCh chan<- *Event, stopChan <-chan bool) error {
	if evCh == nil {
		return ErrNilChan
	}

	req, err := liveReq("GET", uri, nil)
	if err != nil {
		return fmt.Errorf("error getting sse request: %v", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request for %s: %v", uri, err)
	}

	br := bufio.NewReader(res.Body)
	defer res.Body.Close()

	var currEvent *Event
	currEvent = &Event{URI: uri, data: new(bytes.Buffer)}

	for {
		select {
		case <-stopChan:
			return nil
		default:
			bs, err := br.ReadBytes('\n')

			if err != nil && err != io.EOF {
				return err
			}

			if len(bs) < 2 { //newline indicates end of event, emit this one, start populating a new one
				evCh <- currEvent
				currEvent = &Event{URI: uri, data: new(bytes.Buffer)}
				continue
			}

			parseLine(bs, currEvent)

			if err == io.EOF {
				break
			}
		}
	}
}
