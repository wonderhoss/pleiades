package kafka

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/op/go-logging"
)

func TestKafkaPublisher(t *testing.T) {
	logging.InitForTesting(logging.CRITICAL)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kafka Publisher Suite")
}
