package operations

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/metrics"
)

// DeleteClusterOperation describes what the operator does when deleting a cluster
type DeleteClusterOperation struct {
	clusterAccessor   *cluster.Accessor
	clusters          map[string]*cluster.Cluster
	clusterDefinition *v1alpha1.Cassandra
	metricsPoller     *metrics.PrometheusMetrics
}

// Execute performs the operation
func (o *DeleteClusterOperation) Execute() {
	log.Infof("Cassandra cluster definition deleted for cluster: %s.%s", o.clusterDefinition.Namespace, o.clusterDefinition.Name)

	var c *cluster.Cluster
	var ok bool
	if c, ok = o.clusters[fmt.Sprintf("%s.%s", o.clusterDefinition.Namespace, o.clusterDefinition.Name)]; !ok {
		log.Warnf("No record found of deleted cluster %s.%s", o.clusterDefinition.Namespace, o.clusterDefinition.Name)
		return
	}

	delete(o.clusters, c.QualifiedName())
	c.Online = false
	o.metricsPoller.DeleteMetrics(c)

	if err := o.clusterAccessor.DeleteStatefulSetsForCluster(c); err != nil {
		log.Errorf("Error while deleting stateful sets for cluster %s: %v", c.QualifiedName(), err)
	}
	log.Infof("Deleted stateful sets for cluster: %s", c.QualifiedName())

	if err := o.clusterAccessor.DeleteServiceForCluster(c); err != nil {
		log.Errorf("Error while deleting service for cluster %s: %v", c.QualifiedName(), err)
	}
	log.Infof("Deleted headless service for cluster: %s", c.QualifiedName())
	log.Infof("Existing Cassandra cluster removed: %s", c.QualifiedName())
}

func (o *DeleteClusterOperation) String() string {
	return fmt.Sprintf("delete cluster %s", o.clusterDefinition.QualifiedName())
}
