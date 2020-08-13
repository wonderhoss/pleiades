package sse

import (
	"bytes"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var testData1 = []string{
	`:ok`,
	`event: message`,
	`id: [{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`,
	`data: {"$schema":"/mediawiki/recentchange/1.0.0","meta":{"uri":"https://he.wikipedia.org/wiki/%D7%AA%D7%91%D7%A0%D7%99%D7%AA:%D7%A0%D7%AA%D7%95%D7%A0%D7%99_%D7%9E%D7%93%D7%99%D7%A0%D7%95%D7%AA/%D7%A1%D7%9C%D7%95%D7%91%D7%A7%D7%99%D7%94","request_id":"e386ef4b-75f4-46e8-be93-8f3683d30049","id":"9bea80f8-f99c-4b56-93c4-0eb4272bbcb9","dt":"2020-07-31T14:58:47Z","domain":"he.wikipedia.org","stream":"mediawiki.recentchange","topic":"eqiad.mediawiki.recentchange","partition":0,"offset":2603659077},"id":53404707,"type":"edit","namespace":10,"title":"תבנית:נתוני מדינות/סלובקיה","comment":"bot","timestamp":1596207527,"user":"DMbotY","bot":true,"minor":true,"patrolled":true,"length":{"old":4905,"new":4905},"revision":{"old":28682248,"new":28826355},"server_url":"https://he.wikipedia.org","server_name":"he.wikipedia.org","server_script_path":"/w","wiki":"hewiki","parsedcomment":"bot"}`,
}
var testData2 = []string{
	`:ok`,
	`event: message`,
	`id: [{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`,
	`data: {"$schema":"/mediawiki/recentchange/1.0.0",`,
	`data: "meta":{"uri":"https://he.wikipedia.org/wiki/%D7%AA%D7%91%D7%A0%D7%99%D7%AA:%D7%A0%D7%AA%D7%95%D7%A0%D7%99_%D7%9E%D7%93%D7%99%D7%A0%D7%95%D7%AA/%D7%A1%D7%9C%D7%95%D7%91%D7%A7%D7%99%D7%94","request_id":"e386ef4b-75f4-46e8-be93-8f3683d30049","id":"9bea80f8-f99c-4b56-93c4-0eb4272bbcb9","dt":"2020-07-31T14:58:47Z","domain":"he.wikipedia.org","stream":"mediawiki.recentchange","topic":"eqiad.mediawiki.recentchange","partition":0,"offset":2603659077},"id":53404707,"type":"edit","namespace":10,"title":"תבנית:נתוני מדינות/סלובקיה","comment":"bot","timestamp":1596207527,"user":"DMbotY","bot":true,"minor":true,"patrolled":true,"length":{"old":4905,"new":4905},"revision":{"old":28682248,"new":28826355},"server_url":"https://he.wikipedia.org","server_name":"he.wikipedia.org","server_script_path":"/w","wiki":"hewiki","parsedcomment":"bot"}`,
}

var _ = Describe("SSE Receiver", func() {
	Context("Event Parser", func() {
		Context("with single-line data", func() {
			It("tokenizes lines correctly", func() {
				e := &Event{URI: "test", data: new(bytes.Buffer)}
				for _, l := range testData1 {
					parseLine([]byte(l), e)
				}
				Expect(e.Type).Should(Equal("message"))
				Expect(e.ID).Should(Equal(`[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`))
				d, err := ioutil.ReadAll(e.GetData())
				Expect(err).NotTo(HaveOccurred())
				Expect(string(d)).Should(Equal(`{"$schema":"/mediawiki/recentchange/1.0.0","meta":{"uri":"https://he.wikipedia.org/wiki/%D7%AA%D7%91%D7%A0%D7%99%D7%AA:%D7%A0%D7%AA%D7%95%D7%A0%D7%99_%D7%9E%D7%93%D7%99%D7%A0%D7%95%D7%AA/%D7%A1%D7%9C%D7%95%D7%91%D7%A7%D7%99%D7%94","request_id":"e386ef4b-75f4-46e8-be93-8f3683d30049","id":"9bea80f8-f99c-4b56-93c4-0eb4272bbcb9","dt":"2020-07-31T14:58:47Z","domain":"he.wikipedia.org","stream":"mediawiki.recentchange","topic":"eqiad.mediawiki.recentchange","partition":0,"offset":2603659077},"id":53404707,"type":"edit","namespace":10,"title":"תבנית:נתוני מדינות/סלובקיה","comment":"bot","timestamp":1596207527,"user":"DMbotY","bot":true,"minor":true,"patrolled":true,"length":{"old":4905,"new":4905},"revision":{"old":28682248,"new":28826355},"server_url":"https://he.wikipedia.org","server_name":"he.wikipedia.org","server_script_path":"/w","wiki":"hewiki","parsedcomment":"bot"}`))
			})
		})
		Context("with multi-line data", func() {
			It("tokenizes lines correctly", func() {
				e := &Event{URI: "test", data: new(bytes.Buffer)}
				for _, l := range testData2 {
					parseLine([]byte(l), e)
				}
				Expect(e.Type).Should(Equal("message"))
				Expect(e.ID).Should(Equal(`[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`))
				d, err := ioutil.ReadAll(e.GetData())
				Expect(err).NotTo(HaveOccurred())
				Expect(string(d)).Should(Equal(`{"$schema":"/mediawiki/recentchange/1.0.0",
"meta":{"uri":"https://he.wikipedia.org/wiki/%D7%AA%D7%91%D7%A0%D7%99%D7%AA:%D7%A0%D7%AA%D7%95%D7%A0%D7%99_%D7%9E%D7%93%D7%99%D7%A0%D7%95%D7%AA/%D7%A1%D7%9C%D7%95%D7%91%D7%A7%D7%99%D7%94","request_id":"e386ef4b-75f4-46e8-be93-8f3683d30049","id":"9bea80f8-f99c-4b56-93c4-0eb4272bbcb9","dt":"2020-07-31T14:58:47Z","domain":"he.wikipedia.org","stream":"mediawiki.recentchange","topic":"eqiad.mediawiki.recentchange","partition":0,"offset":2603659077},"id":53404707,"type":"edit","namespace":10,"title":"תבנית:נתוני מדינות/סלובקיה","comment":"bot","timestamp":1596207527,"user":"DMbotY","bot":true,"minor":true,"patrolled":true,"length":{"old":4905,"new":4905},"revision":{"old":28682248,"new":28826355},"server_url":"https://he.wikipedia.org","server_name":"he.wikipedia.org","server_script_path":"/w","wiki":"hewiki","parsedcomment":"bot"}`))
			})
		})
	})

	Context("HTTP Client", func() {
		It("produces a correctly configured HTTP Client", func() {
			l, err := liveReq("GET", "http://localhost", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(l.Header.Get("Accept")).Should(Equal("text/event-stream"))
		})
	})
})
