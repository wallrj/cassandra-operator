package metrics

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
)

var _ = Describe("Nodetool Readiness", func() {
	var (
		jolokiaURLProvider *staticURLProvider
		// metricsGatherer    Gatherer
		cluster *cluster.Cluster
	)

	BeforeEach(func() {
		jolokia.responsePrimers = make(map[string]jolokiaResponsePrimer)

		jolokia.returnsNoLiveNodes()
		jolokia.returnsNoUnreachableNodes()
		jolokia.returnsNoJoiningNodes()
		jolokia.returnsNoLeavingNodes()
		jolokia.returnsNoMovingNodes()
		jolokia.returnsRackForNode("racka", "172.16.46.58")
		jolokia.returnsRackForNode("racka", "172.16.101.30")

		jolokiaURLProvider = &staticURLProvider{serverURL}
		// metricsGatherer = NewGatherer(jolokiaURLProvider, &Config{1 * time.Second})

		cluster = aCluster("testcluster", "test")
	})

	Context("Nodetool", func() {
		It("responds to a request", func() {
			// given
			nt := NewNodetool(cluster, jolokiaURLProvider)

			// when
			ready, err := nt.IsLocalNodeReady()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(ready).To(Equal(true))
		})
	})
})
