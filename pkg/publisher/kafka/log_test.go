package kafka

import (
	"github.com/spf13/viper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kafka Crud Logger", func() {

	It("respects verbosity", func() {
		viper.Set("verbose", true)
		a := newCrudLogger()
		Expect(a.debugEnable).Should(BeTrue(), "CrudLogger should be enabled with verbose: true")
		viper.Set("verbose", false)
		b := newCrudLogger()
		Expect(b.debugEnable).Should(BeFalse(), "CrudLogger should not be enabled with verbose: false")
	})

})
