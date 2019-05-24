package metrics

import (
	"log"
	"time"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
)

const (
	defaultLocalUrl = "http://127.0.0.1:7777"
)

type staticURLProvider struct {
	url string
}

func (l *staticURLProvider) UrlFor(*cluster.Cluster) string {
	return l.url
}

type Nodetool struct {
	cluster     *cluster.Cluster
	urlProvider jolokiaURLProvider
}

func NewNodetool(cluster *cluster.Cluster, urlProvider jolokiaURLProvider) *Nodetool {
	return &Nodetool{
		cluster:     cluster,
		urlProvider: urlProvider,
	}
}

func (n *Nodetool) IsLocalNodeReady() (bool, error) {
	gatherer := NewGatherer(n.urlProvider, &Config{
		RequestTimeout: 20 * time.Second,
	})
	status, err := gatherer.GatherMetricsFor(n.cluster)
	if err != nil {
		return false, err
	}
	log.Println("STATUS", status)
	return true, nil
}
