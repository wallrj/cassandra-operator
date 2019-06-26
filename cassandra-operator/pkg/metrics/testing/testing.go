package testing

import (
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
)

// StubbedJolokiaURLProvider is a test stub for JolokiaURLProvider
type StubbedJolokiaURLProvider struct {
	BaseURL string
}

// URLFor returns a fixed Jolokia API URL.
func (p *StubbedJolokiaURLProvider) URLFor(cluster *cluster.Cluster) string {
	return p.BaseURL
}

// JolokiaIsUnavailable causes subsequent calls to URLFor to return an API URL which will not accept connections.
func (p *StubbedJolokiaURLProvider) JolokiaIsUnavailable() {
	p.BaseURL = "localhost:9999"
}
