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

// Newnodetool creates a NodeTool
// urlProvider is optional
// if ommitted a default will be used
func NewNodetool(cluster *cluster.Cluster, urlProvider jolokiaURLProvider) *Nodetool {
	if urlProvider == nil {
		urlProvider = &staticURLProvider{url: defaultLocalUrl}
	}
	return &Nodetool{
		cluster:     cluster,
		urlProvider: urlProvider,
	}
}

func (n *Nodetool) IsNodeReady(host string) (bool, error) {
	gatherer := NewGatherer(n.urlProvider, &Config{
		RequestTimeout: 20 * time.Second,
	})
	status, err := gatherer.GatherMetricsFor(n.cluster)
	if err != nil {
		return false, err
	}
	statusMap := transformClusterStatus(status)
	hostInfo, found := statusMap[host]
	if !found {
		log.Printf("couldn't find status for node: %s", host)
		return false, nil
	}
	log.Println("STATUS", hostInfo)
	return hostInfo.IsUpAndNormal(), nil
}
