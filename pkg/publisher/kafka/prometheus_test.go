package kafka

import (
	"github.com/gargath/pleiades/pkg/sse"
	. "github.com/onsi/ginkgo"
	"github.com/spf13/viper"

	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
)

var _ = Describe("Kafka Client Prometheus Collector", func() {

	It("assembles metrics", func() {
		ch := make(chan (*sse.Event))
		defer close(ch)
		broker := "foo"
		topic := "bar"

		viper.Set("kafka.broker", broker)
		viper.Set("kafka.topic", topic)

		pub, err := NewPublisher(ch)
		Expect(err).NotTo(HaveOccurred())
		p := pub.(*Publisher)
		Expect(p.destination.Topic).Should(Equal(topic))
		Expect(p.destination.Brokers).Should(ContainElement(broker))

		p.currMsgID = `[{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1},{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596550548001}]`
		collector := &PrometheusCollector{Publisher: p}

		dscCh := make(chan (*prometheus.Desc), 100)
		collector.Describe(dscCh)
		close(dscCh)
		dscs := make(map[string]bool)
		for d := range dscCh {
			dscs[d.String()] = true
		}
		Expect(len(dscs)).Should(Equal(6))
		Expect(dscs).Should(HaveKey(`Desc{fqName: "pleiades_kafka_publish_events_total", help: "The total number of messages published to kafka", constLabels: {}, variableLabels: []}`))
		Expect(dscs).Should(HaveKey(`Desc{fqName: "pleiades_kafka_publish_writes_total", help: "The total number of writes performed to kafka", constLabels: {}, variableLabels: []}`))
		Expect(dscs).Should(HaveKey(`Desc{fqName: "pleiades_kafka_writer_errors_total", help: "Total numbers of errors encountered while publishing to kafka", constLabels: {}, variableLabels: []}`))
		Expect(dscs).Should(HaveKey(`Desc{fqName: "pleiades_kafka_publish_write_time_seconds", help: "Time the kafka writer spent writing", constLabels: {}, variableLabels: [agg]}`))
		Expect(dscs).Should(HaveKey(`Desc{fqName: "pleiades_kafka_publish_wait_time_seconds", help: "Time the kafka writer spent waiting", constLabels: {}, variableLabels: [agg]}`))
		Expect(dscs).Should(HaveKey(`Desc{fqName: "pleiades_kafka_publish_lag_milliseconds", help: "Time delay between publish time and timestamp of latest event", constLabels: {}, variableLabels: []}`))
	})
})
