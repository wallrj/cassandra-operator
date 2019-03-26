package dispatcher

import (
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
	"sync"
	"testing"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDispatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Dispatcher Suite", test.CreateParallelReporters("dispatcher"))
}

var _ = Describe("cluster Event dispatcher", func() {
	Context("Event dispatch", func() {
		It("should successfully dispatch a single Event correctly", func() {
			// given
			handler := &counterHandler{}
			dispatcher := New(handler.handle, make(chan struct{}))

			// when
			dispatcher.Dispatch(&Event{Kind: "test", Key: "cluster1", Data: "test"})

			// then
			Eventually(eventProcessedCountFor(handler)).Should(Equal(1))
		})

		It("should successfully dispatch two events for different clusters", func() {
			handler := multiEventHandler{test1Handler: &counterHandler{}, test2Handler: &counterHandler{}}
			dispatcher := New(handler.handle, make(chan struct{}))

			// when
			dispatcher.Dispatch(&Event{Kind: "test1", Key: "cluster1", Data: "test"})
			dispatcher.Dispatch(&Event{Kind: "test2", Key: "cluster2", Data: "test"})

			// then
			Eventually(eventProcessedCountFor(handler.test1Handler)).Should(Equal(1))
			Eventually(eventProcessedCountFor(handler.test2Handler)).Should(Equal(1))
		})
	})

	Context("Event handling", func() {
		Specify("two events bound for the same cluster are handled sequentially", func() {
			handler := &timeRecordingHandler{processedEvents: make(map[string]*timeRecord)}
			dispatcher := New(handler.handle, make(chan struct{}))

			dispatcher.Dispatch(&Event{Kind: "test1", Key: "cluster1", Data: "test"})
			dispatcher.Dispatch(&Event{Kind: "test2", Key: "cluster1", Data: "test"})

			Eventually(func() bool { return handler.eventProcessedCount() == 2 }, 3*time.Second).Should(BeTrue())
			Expect(handler.eventForKey("test1").stopTime).Should(BeTemporally("<", handler.eventForKey("test2").startTime))
		})
	})

	Context("shutting down", func() {
		Specify("should no longer accept events for any consumer when stop channel is closed", func() {
			// given
			stopCh := make(chan struct{})
			handler := &multiEventHandler{test1Handler: &counterHandler{}, test2Handler: &counterHandler{}}
			dispatcher := New(handler.handle, stopCh)

			// when
			dispatcher.Dispatch(&Event{Kind: "test1", Key: "cluster1", Data: "test"})
			dispatcher.Dispatch(&Event{Kind: "test1", Key: "cluster2", Data: "test"})

			time.Sleep(1 * time.Second)
			go close(stopCh)
			time.Sleep(10 * time.Millisecond)

			dispatcher.Dispatch(&Event{Kind: "test1", Key: "cluster1", Data: "test"})
			dispatcher.Dispatch(&Event{Kind: "test1", Key: "cluster2", Data: "test"})

			// then
			time.Sleep(1 * time.Second)
			Expect(handler.test1Handler.eventProcessedCount).Should(Equal(1))
			Expect(handler.test2Handler.eventProcessedCount).Should(Equal(1))
		})
	})
})

func eventProcessedCountFor(handler *counterHandler) func() int {
	return func() int {
		return handler.eventProcessedCount
	}
}

type timeRecord struct {
	startTime time.Time
	stopTime  time.Time
}

type timeRecordingHandler struct {
	processedEvents map[string]*timeRecord
	sync.Mutex
}

func (t *timeRecordingHandler) handle(e *Event) {
	t.Lock()
	defer t.Unlock()

	startTime := time.Now()
	time.Sleep(1 * time.Second)
	stopTime := time.Now()
	t.processedEvents[e.Kind] = &timeRecord{startTime: startTime, stopTime: stopTime}
}

func (t *timeRecordingHandler) eventProcessedCount() int {
	t.Lock()
	defer t.Unlock()
	return len(t.processedEvents)
}

func (t *timeRecordingHandler) eventForKey(key string) *timeRecord {
	t.Lock()
	defer t.Unlock()
	return t.processedEvents[key]
}

type counterHandler struct {
	eventProcessedCount int
}

func (b *counterHandler) handle(e *Event) {
	b.eventProcessedCount++
}

type multiEventHandler struct {
	test1Handler *counterHandler
	test2Handler *counterHandler
}

func (m *multiEventHandler) handle(e *Event) {
	if e.Key == "cluster1" {
		m.test1Handler.handle(e)
	} else if e.Key == "cluster2" {
		m.test2Handler.handle(e)
	}
}
