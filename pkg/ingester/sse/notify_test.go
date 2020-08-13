package sse

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var responseLines = []string{
	`:ok`,
	`event: message`,
	`id: [{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`,
	`data: {"$schema":"/mediawiki/recentchange/1.0.0","meta":{"uri":"https://he.wikipedia.org/wiki/%D7%AA%D7%91%D7%A0%D7%99%D7%AA:%D7%A0%D7%AA%D7%95%D7%A0%D7%99_%D7%9E%D7%93%D7%99%D7%A0%D7%95%D7%AA/%D7%A1%D7%9C%D7%95%D7%91%D7%A7%D7%99%D7%94","request_id":"e386ef4b-75f4-46e8-be93-8f3683d30049","id":"9bea80f8-f99c-4b56-93c4-0eb4272bbcb9","dt":"2020-07-31T14:58:47Z","domain":"he.wikipedia.org","stream":"mediawiki.recentchange","topic":"eqiad.mediawiki.recentchange","partition":0,"offset":2603659077},"id":53404707,"type":"edit","namespace":10,"title":"תבנית:נתוני מדינות/סלובקיה","comment":"bot","timestamp":1596207527,"user":"DMbotY","bot":true,"minor":true,"patrolled":true,"length":{"old":4905,"new":4905},"revision":{"old":28682248,"new":28826355},"server_url":"https://he.wikipedia.org","server_name":"he.wikipedia.org","server_script_path":"/w","wiki":"hewiki","parsedcomment":"bot"}`,
	``,
	`event: message`,
	`id: [{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`,
	`data: {"$schema":"/mediawiki/recentchange/1.0.0",`,
	`data: "meta":{"uri":"https://he.wikipedia.org/wiki/%D7%AA%D7%91%D7%A0%D7%99%D7%AA:%D7%A0%D7%AA%D7%95%D7%A0%D7%99_%D7%9E%D7%93%D7%99%D7%A0%D7%95%D7%AA/%D7%A1%D7%9C%D7%95%D7%91%D7%A7%D7%99%D7%94","request_id":"e386ef4b-75f4-46e8-be93-8f3683d30049","id":"9bea80f8-f99c-4b56-93c4-0eb4272bbcb9","dt":"2020-07-31T14:58:47Z","domain":"he.wikipedia.org","stream":"mediawiki.recentchange","topic":"eqiad.mediawiki.recentchange","partition":0,"offset":2603659077},"id":53404707,"type":"edit","namespace":10,"title":"תבנית:נתוני מדינות/סלובקיה","comment":"bot","timestamp":1596207527,"user":"DMbotY","bot":true,"minor":true,"patrolled":true,"length":{"old":4905,"new":4905},"revision":{"old":28682248,"new":28826355},"server_url":"https://he.wikipedia.org","server_name":"he.wikipedia.org","server_script_path":"/w","wiki":"hewiki","parsedcomment":"bot"}`,
}

var _ = Describe("SSE Consumer", func() {

	var server *httptest.Server
	var resumed bool

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Last-Event-ID") != "" {
				resumed = true
			}
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(200)
			for _, l := range responseLines {
				fmt.Fprintf(w, l)
				fmt.Fprintf(w, "\n")
			}
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "\n")
		}))
	})

	AfterEach(func() {
		resumed = false
		server.Close()
	})

	It("reads and processes events", func() {
		evChan := make(chan *Event)
		clChan := make(chan bool)
		var wg sync.WaitGroup
		events := []Event{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range evChan {
				events = append(events, *e)
			}
		}()
		wg.Add(1)
		go func() { // TODO: This is a bit of a hack. The consumer should terminate naturally when the server connection closes. For some reason, it does not.
			defer wg.Done()
			time.Sleep(2 * time.Second)
			close(clChan)
		}()
		eid, err := Notify(server.URL, "", evChan, clChan)
		close(evChan)
		wg.Wait()
		Expect(err).NotTo(HaveOccurred())
		Expect(eid).Should(Equal(`[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`))
		Expect(len(events)).Should(Equal(2))
		Expect(resumed).Should(BeFalse())
		Expect(events[0].Type).Should(Equal("message"))
		Expect(events[1].Type).Should(Equal("message"))
		Expect(events[0].ID).Should(Equal(`[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`))
		Expect(events[1].ID).Should(Equal(`[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`))
	})

	It("resumes when requested", func() {
		evChan := make(chan *Event)
		clChan := make(chan bool)
		var wg sync.WaitGroup
		events := []Event{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range evChan {
				events = append(events, *e)
			}
		}()
		wg.Add(1)
		go func() { // TODO: This is a bit of a hack. The consumer should terminate naturally when the server connection closes. For some reason, it does not.
			defer wg.Done()
			time.Sleep(2 * time.Second)
			close(clChan)
		}()
		eid, err := Notify(server.URL, "some-event-id", evChan, clChan)
		close(evChan)
		wg.Wait()
		Expect(err).NotTo(HaveOccurred())
		Expect(eid).Should(Equal(`[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`))
		Expect(len(events)).Should(Equal(2))
		Expect(resumed).Should(BeTrue())
		Expect(events[0].Type).Should(Equal("message"))
		Expect(events[1].Type).Should(Equal("message"))
		Expect(events[0].ID).Should(Equal(`[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`))
		Expect(events[1].ID).Should(Equal(`[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`))
	})

	Context("when a server error occurs", func() {
		It("exits cleanly on 404", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(404)
			}))
			defer server.Close()
			evChan := make(chan *Event)
			clChan := make(chan bool)
			var wg sync.WaitGroup
			events := []Event{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				for e := range evChan {
					events = append(events, *e)
				}
			}()
			_, err := Notify(server.URL, "", evChan, clChan)
			close(evChan)
			wg.Wait()
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "non 2xx status code")).To(BeTrue(), "expected error %v to contain 'non 2xx status code'", err)
			Expect(len(events)).Should(Equal(0))
		})

		It("exits cleanly on malformed resposne", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "77")
				w.WriteHeader(200)
			}))
			defer server.Close()
			evChan := make(chan *Event)
			clChan := make(chan bool)
			var wg sync.WaitGroup
			events := []Event{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				for e := range evChan {
					events = append(events, *e)
				}
			}()
			_, err := Notify(server.URL, "", evChan, clChan)
			close(evChan)
			wg.Wait()
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "EOF")).To(BeTrue(), "expected error %v to contain 'EOF'", err)
			Expect(len(events)).Should(Equal(0))
		})
	})
})
