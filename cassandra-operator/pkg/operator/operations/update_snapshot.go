package operations

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

// UpdateSnapshotOperation describes what the operator does when the Snapshot spec is updated for a cluster
type UpdateSnapshotOperation struct {
	cluster         *cluster.Cluster
	clusterAccessor *cluster.Accessor
	newSnapshot     *v1alpha1.Snapshot
	eventRecorder   record.EventRecorder
}

// Execute performs the operation
func (o *UpdateSnapshotOperation) Execute() {
	cassandra := o.cluster.Definition()
	job, err := o.clusterAccessor.FindCronJobForCluster(cassandra, fmt.Sprintf("app=%s", cassandra.SnapshotJobName()))
	if err != nil {
		log.Errorf("Error while retrieving snapshot job for cluster %s: %v", cassandra.QualifiedName(), err)
	}

	if job != nil {
		o.updateSnapshotJob(job)
	}
}

func (o *UpdateSnapshotOperation) updateSnapshotJob(snapshotJob *v1beta1.CronJob) {
	snapshotJob.Spec.Schedule = o.newSnapshot.Schedule
	snapshotJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0] = *o.cluster.CreateSnapshotContainer(o.newSnapshot)
	err := o.clusterAccessor.UpdateCronJob(snapshotJob)
	if err != nil {
		log.Errorf("Error while updating snapshot snapshotJob %s for cluster %s: %v", snapshotJob.Name, o.cluster.QualifiedName(), err)
		return
	}
	o.eventRecorder.Eventf(o.cluster.Definition(), v1.EventTypeNormal, cluster.ClusterSnapshotCreationModificationEvent, "Snapshot creation modified for cluster %s", o.cluster.QualifiedName())
}

func (o *UpdateSnapshotOperation) String() string {
	return fmt.Sprintf("update snapshot schedule for cluster %s", o.cluster.QualifiedName())
}
