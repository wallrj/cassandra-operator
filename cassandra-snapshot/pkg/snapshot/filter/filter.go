package filter

import (
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-snapshot/pkg/nodetool"
	"k8s.io/api/core/v1"
	"strconv"
)

// OutsideRetentionPeriod returns a SnapshotFilter which will provide only Snapshots which are older than the given
// retention cutoff period.
func OutsideRetentionPeriod(pod *v1.Pod, retentionCutoff int64) nodetool.SnapshotFilter {
	return func(snapshots []nodetool.Snapshot) []nodetool.Snapshot {
		var selectedSnapshots []nodetool.Snapshot
		for _, snapshot := range snapshots {
			snapshotTime, err := strconv.Atoi(snapshot.Name)
			if err != nil {
				log.Warnf("Snapshot with name %s on pod %s.%s does not conform to expected naming conventions and will be ignored: %v", snapshot.Name, pod.Namespace, pod.Name, err)
				continue
			}

			if int64(snapshotTime) < retentionCutoff {
				found := false
				for _, selectedSnapshot := range selectedSnapshots {
					found = found || (selectedSnapshot.Name == snapshot.Name && selectedSnapshot.Keyspace == snapshot.Keyspace)
				}

				if !found {
					selectedSnapshots = append(selectedSnapshots, snapshot)
				}
			}
		}

		return selectedSnapshots
	}
}
