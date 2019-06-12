package metrics

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	metricstesting "github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/metrics/testing"
)

var _ = Describe("Nodetool Readiness", func() {
	var (
		jolokiaURLProvider *metricstesting.StubbedJolokiaURLProvider
		cluster            *cluster.Cluster
		nt                 *Nodetool
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

		jolokiaURLProvider = &metricstesting.StubbedJolokiaURLProvider{BaseURL: serverURL}

		cluster = aCluster("testcluster", "test")
		nt = NewNodetool(cluster, jolokiaURLProvider)
	})

	Context("Nodetool", func() {

		It("reports ready when node is healthy", func() {
			// given
			jolokia.returns2LiveNodes()

			// when
			ready, err := nt.IsNodeReady("172.16.46.58")

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(ready).To(Equal(true))
		})

		It("reports unready when node is joining", func() {
			// given
			jolokia.returns2LiveNodes()
			jolokia.returns2JoiningNodes()

			// when
			ready, err := nt.IsNodeReady("172.16.46.58")

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(ready).To(Equal(false))
		})

		It("returns an error if the supplied host is not in the cluster", func() {
			// given
			jolokia.returns2LiveNodes()

			// when
			_, err := nt.IsNodeReady("some.other.host")

			// then
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when jolokia is not available", func() {
			// given
			jolokiaURLProvider.JolokiaIsUnavailable()

			// when
			_, err := nt.IsNodeReady("172.16.46.58")

			// then
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when jolokia returns an error response", func() {
			// given
			jolokia.returnsErrorResponse()

			// when
			ready, err := nt.IsNodeReady("172.16.46.58")

			// then
			Expect(err).To(HaveOccurred())
			Expect(ready).To(Equal(false))
		})

		It("returns an error when jolokia returns an non json response", func() {
			// given
			jolokia.returnsANonJSONResponse()

			// when
			ready, err := nt.IsNodeReady("172.16.46.58")

			// then
			Expect(err).To(HaveOccurred())
			Expect(ready).To(Equal(false))
		})
	})
})
