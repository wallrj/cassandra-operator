package metrics

import (
	"fmt"
	"time"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
)

const (
	defaultLocalURL = "http://127.0.0.1:7777"
)

type staticURLProvider struct {
	url string
}

func (l *staticURLProvider) URLFor(*cluster.Cluster) string {
	return l.url
}

// Nodetool provides a subset of the functions of Cassandra `nodetool`
// as a light weight library.
type Nodetool struct {
	cluster     *cluster.Cluster
	urlProvider JolokiaURLProvider
}

// NewNodetool creates a NodeTool
// urlProvider is optional
// if ommitted a default will be used
func NewNodetool(cluster *cluster.Cluster, urlProvider JolokiaURLProvider) *Nodetool {
	if urlProvider == nil {
		urlProvider = &staticURLProvider{url: defaultLocalURL}
	}
	return &Nodetool{
		cluster:     cluster,
		urlProvider: urlProvider,
	}
}

// IsNodeReady checks whether a particular C* node is UP and NORMAL
func (n *Nodetool) IsNodeReady(host string) (bool, error) {
	gatherer := NewGatherer(n.urlProvider, &Config{
		RequestTimeout: 20 * time.Second,
	})
	clusterStatus, err := gatherer.GatherMetricsFor(n.cluster)
	if err != nil {
		return false, err
	}
	nodeStatus := clusterStatus.NodeStatus(host)
	if nodeStatus == nil {
		return false, fmt.Errorf("node not found: %s", host)
	}
	return nodeStatus.IsUpAndNormal(), nil
}
