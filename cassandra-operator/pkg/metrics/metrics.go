package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
)

// Gatherer defines how metrics will be collected for a cluster
type Gatherer interface {
	GatherMetricsFor(cluster *cluster.Cluster) (*clusterStatus, error)
}

// Config contains options controlling how metrics are fetched
type Config struct {
	RequestTimeout time.Duration
}

// JolokiaURLProvider provides a Jolokia API URL for a Cassandra cluster
type JolokiaURLProvider interface {
	URLFor(*cluster.Cluster) string
}

type jolokiaGatherer struct {
	jolokiaURLProvider JolokiaURLProvider
	httpclient         *http.Client
}

type clusterStatus struct {
	nodeRacks        map[string]string
	liveNodes        []string
	unreachableNodes []string
	joiningNodes     []string
	leavingNodes     []string
	movingNodes      []string
}

// NodeStatus returns the nodeStatus for the supplied host, or nil.
func (cs *clusterStatus) NodeStatus(node string) *nodeStatus {
	return transformClusterStatus(cs)[node]
}

// jolokiaRequest represents the request field in the jolokia response
type jolokiaRequest struct {
	Mbean     string
	Attribute string
	Operation string
}

// JolokiaResponse defines common behaviour across jolokia responses
type jolokiaResponse interface {
	status() uint16
}

// BaseJolokiaResponse represents common fields in a jolokia response
type baseJolokiaResponse struct {
	Request   jolokiaRequest
	Timestamp int64
	Status    uint16
}

func (b *baseJolokiaResponse) status() uint16 {
	return b.Status
}

// MultiValueJolokiaResponse represents a jolokia response with multiple values
type multiValueJolokiaResponse struct {
	baseJolokiaResponse
	Value []string
}

// SingleValueJolokiaResponse represents a jolokia response a single value
type singleValueJolokiaResponse struct {
	baseJolokiaResponse
	Value string
}

// NewGatherer creates a new instance of the Gatherer
func NewGatherer(jolokiaURLProvider JolokiaURLProvider, config *Config) Gatherer {
	return &jolokiaGatherer{
		jolokiaURLProvider: jolokiaURLProvider,
		httpclient:         &http.Client{Timeout: config.RequestTimeout},
	}
}

// GatherMetricsFor retrieves metrics from the jolokia endpoint of a given cluster
func (m *jolokiaGatherer) GatherMetricsFor(cluster *cluster.Cluster) (*clusterStatus, error) {
	mbeanStatusValues, err := m.collectMbeanStatusValuesFor(cluster)
	if err != nil {
		return nil, err
	}

	rackInfo, err := m.collectRackInfoFor(cluster, mbeanStatusValues["LiveNodes"], mbeanStatusValues["UnreachableNodes"])
	if err != nil {
		return nil, err
	}

	return &clusterStatus{
		nodeRacks:        rackInfo,
		liveNodes:        mbeanStatusValues["LiveNodes"],
		unreachableNodes: mbeanStatusValues["UnreachableNodes"],
		joiningNodes:     mbeanStatusValues["JoiningNodes"],
		leavingNodes:     mbeanStatusValues["LeavingNodes"],
		movingNodes:      mbeanStatusValues["MovingNodes"],
	}, nil
}

func (m *jolokiaGatherer) collectRackInfoFor(cluster *cluster.Cluster, liveNodes []string, unreachableNodes []string) (map[string]string, error) {
	clusterJolokiaEndpoint := m.jolokiaURLProvider.URLFor(cluster)

	var allNodes []string
	allNodes = append(allNodes, liveNodes...)
	allNodes = append(allNodes, unreachableNodes...)

	rackInfo := make(map[string]string)
	for _, nodeIP := range allNodes {
		responseHolder := &singleValueJolokiaResponse{}
		err := m.sendRequestToJolokia(fmt.Sprintf("%s/jolokia/exec/org.apache.cassandra.db:type=EndpointSnitchInfo/getRack/%s", clusterJolokiaEndpoint, nodeIP), responseHolder)
		if err != nil {
			return nil, fmt.Errorf("unable to find rack for node %s in cluster %s, %v", nodeIP, cluster.QualifiedName(), err)
		}

		rackInfo[nodeIP] = responseHolder.Value
	}

	return rackInfo, nil
}

func (m *jolokiaGatherer) collectMbeanStatusValuesFor(cluster *cluster.Cluster) (map[string][]string, error) {
	clusterJolokiaEndpoint := m.jolokiaURLProvider.URLFor(cluster)
	mbeanValues := map[string][]string{"LiveNodes": nil, "UnreachableNodes": nil, "JoiningNodes": nil, "LeavingNodes": nil, "MovingNodes": nil}

	for mbean := range mbeanValues {
		rh := &multiValueJolokiaResponse{}
		err := m.sendRequestToJolokia(fmt.Sprintf("%s/jolokia/read/org.apache.cassandra.db:type=StorageService/%s", clusterJolokiaEndpoint, mbean), rh)
		if err != nil {
			return nil, fmt.Errorf("unable to collect metrics for mbean %s for cluster %s, %v", mbean, cluster.QualifiedName(), err)
		}
		mbeanValues[mbean] = rh.Value
	}
	return mbeanValues, nil
}

func (m *jolokiaGatherer) sendRequestToJolokia(jolokiaRequestURL string, responseHolder jolokiaResponse) error {
	resp, err := m.httpclient.Get(jolokiaRequestURL)
	if err != nil {
		return fmt.Errorf("error while retrieving MBean data from URL %s, %v", jolokiaRequestURL, err)
	}

	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while parsing response body for URL %s, %v", jolokiaRequestURL, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("error while retrieving MBean data from URL %s, response body was: %v", jolokiaRequestURL, string(bodyAsBytes))
	}

	if len(bodyAsBytes) > 0 {
		if err := json.Unmarshal(bodyAsBytes, responseHolder); err != nil {
			return fmt.Errorf("error while unmarshalling jolokia response from URL %s. Body %s, %v", jolokiaRequestURL, string(bodyAsBytes), err)
		}
	}

	if responseHolder.status() != 200 {
		return fmt.Errorf("error response returned by jolokia from URL %s. Body %s", jolokiaRequestURL, string(bodyAsBytes))
	}

	return nil
}
