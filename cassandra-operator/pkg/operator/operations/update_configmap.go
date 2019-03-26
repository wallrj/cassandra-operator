package operations

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/operator/operations/adjuster"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

// UpdateCustomConfigOperation describes what the operator does when a configmap is updated for a cluster
type UpdateCustomConfigOperation struct {
	cluster             *cluster.Cluster
	configMap           *v1.ConfigMap
	eventRecorder       record.EventRecorder
	adjuster            *adjuster.Adjuster
	statefulSetAccessor *statefulSetAccessor
}

// Execute performs the operation
func (o *UpdateCustomConfigOperation) Execute() {
	o.eventRecorder.Eventf(o.cluster.Definition(), v1.EventTypeNormal, cluster.ClusterUpdateEvent, "Custom config updated for cluster %s", o.cluster.QualifiedName())
	for _, rack := range o.cluster.Racks() {
		patchChange := o.adjuster.CreateConfigMapHashPatchForRack(&rack, o.configMap)
		if err := o.statefulSetAccessor.patchStatefulSet(o.cluster, patchChange); err != nil {
			log.Errorf("Error while attempting to update rack %s in cluster %s as a result of a custom config change. No further updates will be applied: %v", rack.Name, o.cluster.QualifiedName(), err)
			return
		}
	}
}

func (o *UpdateCustomConfigOperation) String() string {
	return fmt.Sprintf("update custom config for cluster %s", o.cluster.QualifiedName())
}
