package aggregator

import (
	"encoding/json"
	"fmt"

	"github.com/gargath/pleiades/pkg/log"
)

const moduleName = "aggregator"

var (
	logger = log.MustGetLogger(moduleName)
)

// CountersFromEventData parses an event body and generates the Redis counters to increment for it
func CountersFromEventData(data []byte) ([]string, error) {
	var counters = []string{"pleiades_total"}
	var event MediawikiRecentchange
	err := json.Unmarshal(data, &event)
	if err != nil {
		logger.Debugf("failed to parse event data line: %s", string(data))
		return counters, fmt.Errorf("failed to parse event data: %v", err)
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
	}
	return counters, nil
}
