package kafka

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/op/go-logging"
)

func TestAggregator(t *testing.T) {
	logging.InitForTesting(logging.CRITICAL)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kafka Aggregator Suite")
}
