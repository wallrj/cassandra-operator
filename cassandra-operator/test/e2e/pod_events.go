package e2e

import (
	"fmt"
	"github.com/onsi/gomega"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"time"
)

type simplifiedEvent struct {
	timestamp time.Time
	eventType watch.EventType
}

type eventLog interface {
	recordEvent(name string, event simplifiedEvent)
}

type PodEventLog struct {
	events map[string][]simplifiedEvent
}

func WatchPodEvents(namespace, clusterName string) (*PodEventLog, watch.Interface) {
	eventLog := &PodEventLog{make(map[string][]simplifiedEvent)}
	watcher, err := KubeClientset.CoreV1().Pods(namespace).Watch(metaV1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", cluster.OperatorLabel, clusterName)})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	go watchEvents(watcher, eventLog)
	return eventLog, watcher
}

func watchEvents(watcher watch.Interface, events eventLog) {
	for evt := range watcher.ResultChan() {

		switch watchedResource := evt.Object.(type) {
		case *coreV1.Pod:
			podName := watchedResource.Name

			var se simplifiedEvent
			switch evt.Type {
			case watch.Added:
				se = simplifiedEvent{timestamp: watchedResource.CreationTimestamp.Time, eventType: evt.Type}
			case watch.Deleted:
				se = simplifiedEvent{timestamp: watchedResource.DeletionTimestamp.Time, eventType: evt.Type}
			default:
				continue
			}
			events.recordEvent(podName, se)
		case *v1alpha1.Cassandra:
			events.recordEvent(watchedResource.Name, simplifiedEvent{timestamp: watchedResource.CreationTimestamp.Time, eventType: evt.Type})
		default:
			continue
		}
	}
}

func (e *PodEventLog) PodsRecreatedOneAfterTheOther(pods ...string) (bool, error) {
	for i := 0; i < len(pods)-1; i++ {
		firstPod := pods[i]
		secondPod := pods[i+1]

		lastCreationTimeForFirstPod, err := e.lastCreationTimeForPod(firstPod)
		if err != nil {
			return false, err
		}

		deletionTimeForSecondPod, err := e.deletionTimeForPod(secondPod)
		if err != nil {
			return false, err
		}

		if deletionTimeForSecondPod.Before(lastCreationTimeForFirstPod) {
			return false, fmt.Errorf("second pod was deleted before first pod was recreated. Second pod deleted at: %v, first pod last created at: %v", deletionTimeForSecondPod, lastCreationTimeForFirstPod)
		}
	}
	return true, nil
}

func (e *PodEventLog) PodsStartedEventCount(pod string) int {
	return len(e.findEventsOfType(pod, watch.Added))
}

func (e *PodEventLog) recordEvent(podName string, event simplifiedEvent) {
	if _, ok := e.events[podName]; !ok {
		e.events[podName] = []simplifiedEvent{}
	}
	e.events[podName] = append(e.events[podName], event)
}

func (e *PodEventLog) deletionTimeForPod(podName string) (time.Time, error) {
	return e.findLastEventTime(podName, watch.Deleted)
}

func (e *PodEventLog) lastCreationTimeForPod(podName string) (time.Time, error) {
	return e.findLastEventTime(podName, watch.Added)
}

func (e *PodEventLog) findLastEventTime(podName string, eventType watch.EventType) (time.Time, error) {
	podEvent, err := e.findLastEventOfType(podName, eventType)
	if err != nil {
		return time.Time{}, err
	}
	return podEvent.timestamp, nil
}

func (e *PodEventLog) findLastEventOfType(podName string, eventType watch.EventType) (*simplifiedEvent, error) {
	podEvents := e.findEventsOfType(podName, eventType)
	if len(podEvents) == 0 {
		return nil, fmt.Errorf("no events found for pod: %s", podName)
	}

	for i := len(podEvents) - 1; i >= 0; i-- {
		if podEvents[i].eventType == eventType {
			return &podEvents[i], nil
		}
	}
	return nil, fmt.Errorf("no events of type %s found for pod: %s", eventType, podName)
}

func (e *PodEventLog) findEventsOfType(podName string, eventType watch.EventType) []simplifiedEvent {
	var podEvents []simplifiedEvent
	for _, event := range e.events[podName] {
		if event.eventType == eventType {
			podEvents = append(podEvents, event)
		}
	}
	return podEvents
}
