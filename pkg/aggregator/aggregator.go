package aggregator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/gargath/pleiades/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const moduleName = "aggregator"

var (
	logger = log.MustGetLogger(moduleName)

	timeStampRegExp = regexp.MustCompile(`"timestamp":([0-9]+).*`)

	msgLag = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "pleiades_aggregator_message_lag_milliseconds",
			Help:    "Age of messages processed",
			Buckets: []float64{1000, 5000, 15000, 60000, 600000, 7200000},
		},
	)
)

// CountersFromEventData parses an event body and generates the Redis counters to increment for it
func CountersFromEventData(data []byte) ([]string, int64, error) { //TODO: This should return a set of counters and increments to allow for more than just +1
	var lendiff int64 = 0
	var counters = []string{"pleiades_total"}
	var event MediawikiRecentchange
	err := json.Unmarshal(data, &event)
	if err != nil {
		logger.Debugf("failed to parse event data line: %s", string(data))
		return counters, 0, fmt.Errorf("failed to parse event data: %v", err)
	}
	if event.Wiki != "" {
		counters = append(counters, "pleiades_wiki_"+event.Wiki)
	} else {
		logger.Info("Encountered event without a Wiki: %+v", event)
	}
	if event.Type != "" {
		counters = append(counters, "pleiades_type_"+event.Type)
	} else {
		logger.Info("Encountered event without Type: %+v", event)
	}
	if event.Bot {
		counters = append(counters, "pleiades_bot")
	}
	if event.Minor {
		counters = append(counters, "pleiades_minor")
	}
	if event.Length != nil {
		if event.Length.Old < event.Length.New {
			counters = append(counters, "pleiades_length_inc")
		} else {
			counters = append(counters, "pleiades_length_dec")
		}
		lendiff = event.Length.New - event.Length.Old
	}
	return counters, lendiff, nil
}

// RecordLag parses the timestamp from a event ID and observes the lag as Prometheus metrics
func RecordLag(id string) {
	timeStamp, err := parseTimestamp(id)
	if err != nil {
		logger.Errorf("Error parsing event ID: %v", err)
	}
	lag := (time.Now().UnixNano() - (timeStamp * 1000000)) / 1000000 // timestamp is ms, so convert to ns, subtract from UnixNano(), then convert back to ms
	msgLag.Observe(float64(lag))
}

func parseTimestamp(id string) (int64, error) {
	match := timeStampRegExp.FindStringSubmatch(id)
	if len(match) < 2 {
		return 0, fmt.Errorf("Event ID %s has no timestamp", id)
	}
	timeStamp, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse timestamp from ID %s: %v", id, err)
	}
	return timeStamp, nil
}
