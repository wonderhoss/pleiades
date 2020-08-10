package aggregator

import (
	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
)

var _ = Describe("Aggregator Parser", func() {

	It("correctly parses timestamp from event ID", func() {
		ids := map[float64]string{
			1597056638001: `[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1597056638001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`,
			1597056638002: `[{"topic":"eqiad.mediawiki.recentchange","timestamp":1597056638002,"partition":0},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`,
			0:             `[{"topic":"eqiad.mediawiki.recentchange","offset":01,"partition":0},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`,
			1597056638004: `[{"timestamp":1597056638004, "topic":"eqiad.mediawiki.recentchange","partition":0},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`,
		}

		for k, v := range ids {
			ts, err := parseTimestamp(v)
			Expect(ts).Should(Equal(float64(k)))
			if k == float64(0) {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		}
	})
})
