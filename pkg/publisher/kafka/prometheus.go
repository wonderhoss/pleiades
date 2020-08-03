package kafka

import (
	"github.com/prometheus/client_golang/prometheus"
	kafka "github.com/segmentio/kafka-go"
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
)

// PrometheusCollector reports stats from the kafka client to Prometheus
type PrometheusCollector struct {
	Writer *kafka.Writer
}

// Describe implements the Collector's Describe method
func (k PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(k, ch)
}

// Collect implements the Collector's Collect method
func (k PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	stats := k.Writer.Stats()

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
}
