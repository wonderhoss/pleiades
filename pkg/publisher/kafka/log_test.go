package kafka

import (
	"github.com/op/go-logging"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kafka Crud Logger", func() {

	It("respects verbosity", func() {
		logging.SetLevel(logging.DEBUG, "kafka-client")
		a := newCrudLogger()
		Expect(a.debugEnable).Should(BeTrue(), "CrudLogger should be enabled with verbose: true")
		logging.SetLevel(logging.INFO, "kafka-client")
		b := newCrudLogger()
		Expect(b.debugEnable).Should(BeFalse(), "CrudLogger should not be enabled with verbose: false")
	})

})
