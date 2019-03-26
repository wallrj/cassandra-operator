package metrics

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node status derivation", func() {
	Context("Cluster metrics have been collected", func() {
		It("has one live and normal node", func() {
			nodeStatuses := transformClusterStatus(&clusterStatus{liveNodes: []string{"10.0.0.1"}})
			Expect(nodeStatuses["10.0.0.1"].livenessLabel()).To(Equal("up"))
			Expect(nodeStatuses["10.0.0.1"].stateLabel()).To(Equal("normal"))
		})

		It("has one live and joining node", func() {
			nodeStatuses := transformClusterStatus(&clusterStatus{
				liveNodes:    []string{"10.0.0.1"},
				joiningNodes: []string{"10.0.0.1"},
			})
			Expect(nodeStatuses["10.0.0.1"].livenessLabel()).To(Equal("up"))
			Expect(nodeStatuses["10.0.0.1"].stateLabel()).To(Equal("joining"))
		})

		It("has one live and moving node", func() {
			nodeStatuses := transformClusterStatus(&clusterStatus{
				liveNodes:   []string{"10.0.0.1"},
				movingNodes: []string{"10.0.0.1"},
			})
			Expect(nodeStatuses["10.0.0.1"].livenessLabel()).To(Equal("up"))
			Expect(nodeStatuses["10.0.0.1"].stateLabel()).To(Equal("moving"))
		})

		It("has one live and leaving node", func() {
			nodeStatuses := transformClusterStatus(&clusterStatus{
				liveNodes:    []string{"10.0.0.1"},
				leavingNodes: []string{"10.0.0.1"},
			})
			Expect(nodeStatuses["10.0.0.1"].livenessLabel()).To(Equal("up"))
			Expect(nodeStatuses["10.0.0.1"].stateLabel()).To(Equal("leaving"))
		})

		It("has one down and normal node", func() {
			nodeStatuses := transformClusterStatus(&clusterStatus{
				unreachableNodes: []string{"10.0.0.1"},
			})
			Expect(nodeStatuses["10.0.0.1"].livenessLabel()).To(Equal("down"))
			Expect(nodeStatuses["10.0.0.1"].stateLabel()).To(Equal("normal"))
		})

		It("has one live and normal node and one live and joining node", func() {
			nodeStatuses := transformClusterStatus(&clusterStatus{
				liveNodes:    []string{"10.0.0.1", "10.0.0.2"},
				joiningNodes: []string{"10.0.0.2"},
			})
			Expect(nodeStatuses["10.0.0.1"].livenessLabel()).To(Equal("up"))
			Expect(nodeStatuses["10.0.0.1"].stateLabel()).To(Equal("normal"))

			Expect(nodeStatuses["10.0.0.2"].livenessLabel()).To(Equal("up"))
			Expect(nodeStatuses["10.0.0.2"].stateLabel()).To(Equal("joining"))
		})

		It("reports all liveness/state combinations apart from up/normal as not applicable for a live and normal node", func() {
			nodeStatuses := transformClusterStatus(&clusterStatus{liveNodes: []string{"10.0.0.1"}})

			Expect(nodeStatuses["10.0.0.1"].unapplicableLabelPairs()).To(Equal([]labelPair{
				{"up", "joining"},
				{"up", "leaving"},
				{"up", "moving"},
				{"down", "normal"},
				{"down", "joining"},
				{"down", "leaving"},
				{"down", "moving"},
			}))
		})

		It("reports correct unapplicable state combinations for two nodes in different states", func() {
			nodeStatuses := transformClusterStatus(&clusterStatus{
				liveNodes:        []string{"10.0.0.1"},
				unreachableNodes: []string{"10.0.0.2"},
				joiningNodes:     []string{"10.0.0.1"},
			})

			Expect(nodeStatuses["10.0.0.1"].unapplicableLabelPairs()).To(Equal([]labelPair{
				{"up", "normal"},
				{"up", "leaving"},
				{"up", "moving"},
				{"down", "normal"},
				{"down", "joining"},
				{"down", "leaving"},
				{"down", "moving"},
			}))

			Expect(nodeStatuses["10.0.0.2"].unapplicableLabelPairs()).To(Equal([]labelPair{
				{"up", "normal"},
				{"up", "joining"},
				{"up", "leaving"},
				{"up", "moving"},
				{"down", "joining"},
				{"down", "leaving"},
				{"down", "moving"},
			}))
		})
	})
})
