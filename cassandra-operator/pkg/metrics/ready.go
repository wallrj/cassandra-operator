package metrics

import (
	"log"
	"time"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
)

type localGatherer struct {
}

func (l *localGatherer) UrlFor(*cluster.Cluster) string {
	return "http://127.0.0.1:7777"
}

type Nodetool struct {
	cluster *cluster.Cluster
}

func NewNodetool(cluster *cluster.Cluster) *Nodetool {
	return &Nodetool{cluster: cluster}
}

func (n *Nodetool) IsLocalNodeReady() (bool, error) {
	j := &localGatherer{}
	gatherer := NewGatherer(j, &Config{
		RequestTimeout: 20 * time.Second,
	})
	status, err := gatherer.GatherMetricsFor(n.cluster)
	if err != nil {
		return false, err
	}
	log.Println(status)
	return true, nil
}
