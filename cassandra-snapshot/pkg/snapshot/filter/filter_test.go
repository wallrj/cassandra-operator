package filter

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/cassandra-operator/cassandra-snapshot/pkg/nodetool"
	"github.com/sky-uk/cassandra-operator/cassandra-snapshot/test"
	"k8s.io/api/core/v1"
	"strconv"
	"testing"
	"time"
)

func TestFilter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Filter Unit Tests", test.CreateReporters("filter"))
}

var _ = Describe("filtering of snapshots", func() {
	var testStartTime int64
	var filterUnderTest nodetool.SnapshotFilter

	BeforeEach(func() {
		pod := &v1.Pod{}
		pod.Name = "test"
		pod.Namespace = "test"
		testStartTime = time.Now().Unix()
		filterUnderTest = OutsideRetentionPeriod(pod, testStartTime-int64(time.Hour.Seconds()))
	})

	It("should return an empty slice of snapshots when given an empty slice", func() {
		// given
		var snapshotsIn []nodetool.Snapshot

		// when
		snapshotsOut := filterUnderTest(snapshotsIn)

		// then
		Expect(snapshotsOut).To(HaveLen(0))
	})

	It("should filter out all snapshots when all supplied snapshots are within the retention period", func() {
		// given
		now := int(time.Now().Unix())

		snapshotsIn := []nodetool.Snapshot{
			{Name: strconv.Itoa(now), Keyspace: "a", ColumnFamily: "a"},
			{Name: strconv.Itoa(now - 3600), Keyspace: "a", ColumnFamily: "a"},
			{Name: strconv.Itoa(now - 600), Keyspace: "a", ColumnFamily: "a"},
		}

		// when
		snapshotsOut := filterUnderTest(snapshotsIn)

		// then
		Expect(snapshotsOut).To(HaveLen(0))
	})

	It("should filter out no snapshots when all supplied snapshots are outside the retention period", func() {
		// given
		now := int(time.Now().Unix())

		snapshotsIn := []nodetool.Snapshot{
			{Name: strconv.Itoa(now - 3601), Keyspace: "a", ColumnFamily: "a"},
			{Name: strconv.Itoa(now - 7200), Keyspace: "a", ColumnFamily: "a"},
		}

		// when
		snapshotsOut := filterUnderTest(snapshotsIn)

		// then
		Expect(snapshotsOut).To(ConsistOf(
			nodetool.Snapshot{Name: strconv.Itoa(now - 3601), Keyspace: "a", ColumnFamily: "a"},
			nodetool.Snapshot{Name: strconv.Itoa(now - 7200), Keyspace: "a", ColumnFamily: "a"},
		))
	})

	It("should filter out only the snapshots outside the retention period when give a mixed list of snapshots", func() {
		// given
		now := int(time.Now().Unix())

		snapshotsIn := []nodetool.Snapshot{
			{Name: strconv.Itoa(now), Keyspace: "a", ColumnFamily: "a"},
			{Name: strconv.Itoa(now - 3600), Keyspace: "a", ColumnFamily: "a"},
			{Name: strconv.Itoa(now - 600), Keyspace: "a", ColumnFamily: "a"},
			{Name: strconv.Itoa(now - 3601), Keyspace: "a", ColumnFamily: "a"}, // outside retention period
			{Name: strconv.Itoa(now - 7200), Keyspace: "a", ColumnFamily: "a"}, // outside retention period
		}

		// when
		snapshotsOut := filterUnderTest(snapshotsIn)

		// then
		Expect(snapshotsOut).To(ConsistOf(
			nodetool.Snapshot{Name: strconv.Itoa(now - 3601), Keyspace: "a", ColumnFamily: "a"},
			nodetool.Snapshot{Name: strconv.Itoa(now - 7200), Keyspace: "a", ColumnFamily: "a"},
		))
	})

	It("should not return more than one snapshot where the snapshot name and keyspace are the same", func() {
		// given
		snapshotsIn := []nodetool.Snapshot{
			{Name: "1", Keyspace: "a", ColumnFamily: "a"},
			{Name: "1", Keyspace: "a", ColumnFamily: "b"},
			{Name: "1", Keyspace: "b", ColumnFamily: "a"},
			{Name: "2", Keyspace: "a", ColumnFamily: "a"},
		}

		// when
		snapshotsOut := filterUnderTest(snapshotsIn)

		// then
		Expect(snapshotsOut).To(ConsistOf(
			nodetool.Snapshot{Name: "1", Keyspace: "a", ColumnFamily: "a"},
			nodetool.Snapshot{Name: "1", Keyspace: "b", ColumnFamily: "a"},
			nodetool.Snapshot{Name: "2", Keyspace: "a", ColumnFamily: "a"},
		))
	})

	It("should filter out snapshots whose name is not numeric", func() {
		// given
		snapshotsIn := []nodetool.Snapshot{
			{Name: "x", Keyspace: "a", ColumnFamily: "a"},
			{Name: "x!", Keyspace: "a", ColumnFamily: "a"},
			{Name: "x.", Keyspace: "a", ColumnFamily: "a"},
		}

		// when
		snapshotsOut := filterUnderTest(snapshotsIn)

		// then
		Expect(snapshotsOut).To(HaveLen(0))
	})
})
