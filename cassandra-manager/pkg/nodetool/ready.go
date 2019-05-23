package nodetool

import (
	"time"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/metrics"
)

type localGatherer struct {
}

func (l *localGatherer) UrlFor(*cluster.Cluster) string {
	return "http://127.0.0.1:7777"
}

func IsLocalNodeReady() bool {
	j := &localGatherer{}
	gatherer := metrics.NewGatherer(j, &metrics.Config{RequestTimeout: 20 * time.Second})
	status := gatherer.GatherMetricsFor(&cluster.Cluster{})
	return true
}
