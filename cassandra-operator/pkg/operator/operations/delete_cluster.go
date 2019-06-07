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
}

func (o *DeleteClusterOperation) String() string {
	return fmt.Sprintf("delete cluster %s", o.clusterDefinition.QualifiedName())
}
