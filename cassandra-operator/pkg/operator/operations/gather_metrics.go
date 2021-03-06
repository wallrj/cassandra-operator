package operations

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/metrics"
)

// GatherMetricsOperation describes what the operator does when gathering metrics
type GatherMetricsOperation struct {
	metricsPoller *metrics.PrometheusMetrics
	cluster       *cluster.Cluster
}

// Execute performs the operation
func (o *GatherMetricsOperation) Execute() {
	log.Debugf("Processing request to update metrics for %s", o.cluster.QualifiedName())
	o.metricsPoller.UpdateMetrics(o.cluster)
}

func (o *GatherMetricsOperation) String() string {
	return fmt.Sprintf("gather metrics for cluster %s", o.cluster.QualifiedName())
}
