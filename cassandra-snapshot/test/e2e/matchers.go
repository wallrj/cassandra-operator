package e2e

import (
	"fmt"
	"github.com/onsi/gomega/types"
	"strconv"
	"time"
)

func BeForKeyspace(keyspace string) *SnapshotMatcher {
	return &SnapshotMatcher{keyspace: keyspace}
}

type SnapshotMatcher struct {
	snapshotEarliestTime time.Time
	snapshotLatestTime   time.Time
	keyspace             string
}

func (m *SnapshotMatcher) AndWithinTimeRange(earliestTime time.Time, latestTime time.Time) types.GomegaMatcher {
	m.snapshotEarliestTime = earliestTime
	m.snapshotLatestTime = latestTime
	return m
}

func (m *SnapshotMatcher) Match(actual interface{}) (success bool, err error) {
	actualSnapshot := actual.(Snapshot)
	actualTimestamp, err := strconv.Atoi(actualSnapshot.Name)
	if err != nil {
		return false, err
	}
	snapshotTime := time.Unix(int64(actualTimestamp), 0)

	return snapshotTime.Unix() >= m.snapshotEarliestTime.Unix() &&
		snapshotTime.Unix() <= m.snapshotLatestTime.Unix() &&
		actualSnapshot.Keyspace == m.keyspace, nil

}

func (m *SnapshotMatcher) FailureMessage(actual interface{}) (message string) {
	actualSnapshots := actual.([]Snapshot)
	return fmt.Sprintf("Snapshot %v to be found in snapshot list: %v", m, actualSnapshots)
}

func (m *SnapshotMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	actualSnapshots := actual.([]Snapshot)
	return fmt.Sprintf("Snapshot %v not to be found in snapshot list: %v", m, actualSnapshots)
}
