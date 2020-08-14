// This package includes code Copyright (c) 2015 Andrew Stuart,
// licensed under MIT

package sse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gargath/pleiades/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const moduleName = "sse"

var (
	succChan = make(chan (*http.Response))
	errChan  = make(chan (error))
)

// TODO: Rework metrics to ensure they only register when correct personality is running.
// Currently aggregators expose these, too.
var (
	lastEventID    string
	eventsReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pleiades_recv_events_total",
			Help: "The total number of events received"})
	linesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pleiades_recv_event_lines_total",
			Help: "Total numbers of lines read from server",
		},
		[]string{"type"})
	recvErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pleiades_recv_errors_total",
			Help: "Total numbers of errors encountered during events receive",
		},
		[]string{"type"})
	logger = log.MustGetLogger(moduleName)
)

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
			linesReceived.WithLabelValues("comment").Inc()
			return
		}
		linesReceived.WithLabelValues("unknown").Inc()
		logger.Warningf("WARN: encountered non-SSE-compliant line in server response: %s", string(bs))
	}
	switch string(spl[0]) {
	case iName:
		linesReceived.WithLabelValues("id").Inc()
		e := string(bytes.TrimSpace(spl[1]))
		currEvent.ID = e
		lastEventID = e
	case eName:
		linesReceived.WithLabelValues("event").Inc()
		currEvent.Type = string(bytes.TrimSpace(spl[1]))
	case dName:
		linesReceived.WithLabelValues("data").Inc()
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
func Notify(uri string, resumeID string, evCh chan<- *Event, stopChan <-chan bool) (string, error) {
	client := &http.Client{}
	if evCh == nil {
		return lastEventID, ErrNilChan
	}

	req, err := liveReq("GET", uri, nil)
	if err != nil {
		logger.Errorf("Error creating HTTP request: %v", err)
		return lastEventID, fmt.Errorf("error getting sse request: %v", err)
	}
	if resumeID != "" {
		logger.Infof("Requesting subscription to resume from %s", resumeID)
		req.Header.Set("Last-Event-ID", resumeID)
	} else {
		logger.Info("Starting new subscription")
	}
	var res *http.Response

	go func() {
		response, responseError := client.Do(req)
		if responseError != nil {
			errChan <- responseError
		}
		succChan <- response
	}()
	select {
	case err := <-errChan:
		logger.Errorf("Error performing HTTP request for %s: %v", uri, err)
		return lastEventID, fmt.Errorf("error performing request for %s: %v", uri, err)
	case <-time.After(60 * time.Second):
		recvErrors.WithLabelValues("request_timeout").Inc()
		return lastEventID, fmt.Errorf("timeout performing HTTP request")
	case resp := <-succChan:
		if resp == nil {
			return lastEventID, fmt.Errorf("unknown error reading HTTP response")
		}
		if resp.StatusCode > 299 {
			logger.Errorf("Server at %s responded %s", uri, resp.StatusCode)
			return lastEventID, fmt.Errorf("non 2xx status code from request for %s: %d", uri, resp.StatusCode)
		}
		res = resp
	}

	br := bufio.NewReader(res.Body)

	defer res.Body.Close()

	var currEvent *Event
	currEvent = &Event{URI: uri, data: new(bytes.Buffer)}

	for {
		select {
		case <-stopChan:
			logger.Debug("SSE consumer stopped")
			return lastEventID, nil
		default:
			lineChan := make(chan ([]byte))
			errChan := make(chan (error))
			go func() {
				bodyBytes, err := br.ReadBytes('\n')
				if err != nil {
					errChan <- err
				} else {
					lineChan <- bodyBytes
				}
			}()
			var bs []byte
			select {
			case lineBytes := <-lineChan:
				bs = make([]byte, len(lineBytes))
				copy(bs, lineBytes)
			case rderr := <-errChan:
				if rderr != io.EOF {
					recvErrors.WithLabelValues("read_error").Inc()
					return lastEventID, fmt.Errorf("error reading from response body: %v", rderr)
				}
				recvErrors.WithLabelValues("eof").Inc()
				logger.Warning("encountered EOF while reading server response - consumer terminating")
				return lastEventID, nil
			case <-time.After(60 * time.Second):
				logger.Warning("timeout reading from response body")
				recvErrors.WithLabelValues("body_read_timeout").Inc()
				return lastEventID, fmt.Errorf("timeout while reading from response body")
			}

			if len(bs) < 2 { //newline indicates end of event, emit this one, start populating a new one
				if currEvent.ID != "" || currEvent.Type != "" || currEvent.data.Len() > 0 {
					eventsReceived.Inc()
					evCh <- currEvent
					currEvent = &Event{URI: uri, data: new(bytes.Buffer)}
				}
				continue
			}

			parseLine(bs, currEvent)
		}
	}
}
