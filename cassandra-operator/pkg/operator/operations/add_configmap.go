package operations

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/operator/operations/adjuster"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

// AddCustomConfigOperation describes what the operator does when a configmap is added
type AddCustomConfigOperation struct {
	cluster             *cluster.Cluster
	configMap           *v1.ConfigMap
	eventRecorder       record.EventRecorder
	adjuster            *adjuster.Adjuster
	statefulSetAccessor *statefulSetAccessor
}

// Execute performs the operation
func (o *AddCustomConfigOperation) Execute() {
	cassandra := o.cluster.Definition()
	o.eventRecorder.Eventf(cassandra, v1.EventTypeNormal, cluster.ClusterUpdateEvent, "Custom config created for cluster %s", cassandra.QualifiedName())
	for _, rack := range o.cluster.Racks() {
		err := o.statefulSetAccessor.updateStatefulSet(o.cluster, o.configMap, &rack, o.cluster.AddCustomConfigVolumeToStatefulSet)
		if err != nil {
			log.Errorf("unable to add custom configMap to statefulSet for rack %s in cluster %s: %v. Other racks will not be updated", rack.Name, cassandra.QualifiedName(), err)
			return
		}
	}
}

func (o *AddCustomConfigOperation) String() string {
	return fmt.Sprintf("add custom config for cluster %s", o.cluster.QualifiedName())
}
