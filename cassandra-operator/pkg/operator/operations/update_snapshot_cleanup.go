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

// UpdateSnapshotCleanupOperation describes what the operator does when the retention policy is updated for a cluster
type UpdateSnapshotCleanupOperation struct {
	cluster         *cluster.Cluster
	clusterAccessor *cluster.Accessor
	newSnapshot     *v1alpha1.Snapshot
	eventRecorder   record.EventRecorder
}

// Execute performs the operation
func (o *UpdateSnapshotCleanupOperation) Execute() {
	cassandra := o.cluster.Definition()
	job, err := o.clusterAccessor.FindCronJobForCluster(cassandra, fmt.Sprintf("app=%s", cassandra.SnapshotCleanupJobName()))
	if err != nil {
		log.Errorf("Error while retrieving snapshot cleanup job for cluster %s: %v", cassandra.QualifiedName(), err)
	}

	if job != nil {
		o.updateSnapshotCleanupJob(job)
	}
}

func (o *UpdateSnapshotCleanupOperation) updateSnapshotCleanupJob(job *v1beta1.CronJob) {
	job.Spec.Schedule = o.newSnapshot.RetentionPolicy.CleanupSchedule
	job.Spec.JobTemplate.Spec.Template.Spec.Containers[0] = *o.cluster.CreateSnapshotCleanupContainer(o.newSnapshot)
	err := o.clusterAccessor.UpdateCronJob(job)
	if err != nil {
		log.Errorf("Error while updating snapshot cleanup job %s for cluster %s: %v", job.Name, o.cluster.QualifiedName(), err)
		return
	}
	o.eventRecorder.Eventf(o.cluster.Definition(), v1.EventTypeNormal, cluster.ClusterSnapshotCleanupModificationEvent, "Snapshot cleanup modified for cluster %s", o.cluster.QualifiedName())
}

func (o *UpdateSnapshotCleanupOperation) String() string {
	return fmt.Sprintf("update snapshot cleanup schedule for cluster %s", o.cluster.QualifiedName())
}
