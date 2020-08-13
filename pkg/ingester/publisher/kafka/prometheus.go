package kafka

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	messages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pleiades_kafka_publish_events_total",
		Help: "The total number of messages published to kafka",
	})

	writes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pleiades_kafka_publish_writes_total",
		Help: "The total number of writes performed to kafka",
	})

	writeErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pleiades_kafka_writer_errors_total",
		Help: "Total numbers of errors encountered while publishing to kafka",
	})

	kafkaWriteTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pleiades_kafka_publish_write_time_seconds",
		Help: "Time the kafka writer spent writing",
	},
		[]string{"agg"},
	)

	kafkaWaitTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pleiades_kafka_publish_wait_time_seconds",
		Help: "Time the kafka writer spent waiting",
	},
		[]string{"agg"},
	)

	kafkaLag = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pleiades_kafka_publish_lag_milliseconds",
		Help: "Time delay between publish time and timestamp of latest event",
	})
)

// PrometheusCollector reports stats from the kafka client to Prometheus
type PrometheusCollector struct {
	Publisher *Publisher
}

// Describe implements the Collector's Describe method
func (k PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(k, ch)
}

// Collect implements the Collector's Collect method
func (k PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	stats := k.Publisher.w.Stats()

	messages.Add(float64(stats.Messages))
	ch <- messages

	writes.Add(float64(stats.Writes))
	ch <- writes

	writeErrors.Add(float64(stats.Errors))
	ch <- writeErrors

	kafkaWriteTime.WithLabelValues("min").Set(stats.WriteTime.Min.Seconds())
	kafkaWriteTime.WithLabelValues("max").Set(stats.WriteTime.Max.Seconds())
	kafkaWriteTime.WithLabelValues("avg").Set(stats.WriteTime.Avg.Seconds())
	kafkaWriteTime.Collect(ch)

	kafkaWaitTime.WithLabelValues("min").Set(stats.WaitTime.Min.Seconds())
	kafkaWaitTime.WithLabelValues("max").Set(stats.WaitTime.Max.Seconds())
	kafkaWaitTime.WithLabelValues("avg").Set(stats.WaitTime.Avg.Seconds())
	kafkaWaitTime.Collect(ch)

	if k.Publisher.currMsgID == "" {
		return
	}

	now := time.Now().UnixNano() / 1000000
	msgTimestamp, err := tStampFromID(k.Publisher.currMsgID)
	logger.Debugf("Time now is %d, last Timestamp was %d, lag is thus %d ms", now, msgTimestamp, now-msgTimestamp)
	if err != nil {
		logger.Errorf("Error parsing timestamp from event ID %s: %v", k.Publisher.currMsgID, err)
	}
	lag := now - msgTimestamp
	kafkaLag.Set(float64(lag))
	ch <- kafkaLag
}

func tStampFromID(id string) (int64, error) {
	tokens := strings.SplitAfter(id, "timestamp\":")
	if len(tokens) < 2 {
		return 0, fmt.Errorf("'timestamp' string not found")
	}
	stampString := strings.Split(tokens[1], "}")[0]
	stamp, err := strconv.ParseInt(stampString, 10, 64)
	if err != nil {
		return 0, err
	}
	return stamp, nil
}
