package e2e

import (
	"github.com/onsi/ginkgo"
	"strings"
)

func ParallelTestBeforeSuite(createClustersOnce func() []TestCluster, runOnceOnEachNode func(clusterNames []string)) bool {
	return ginkgo.SynchronizedBeforeSuite(func() []byte {
		DeleteCassandraResourcesInNamespace(Namespace)
		testClusters := createClustersOnce()
		clusterNames := []string{}
		for _, cluster := range testClusters {
			clusterNames = append(clusterNames, cluster.Name)
		}
		return []byte(strings.Join(clusterNames, "|"))
	}, func(data []byte) {
		clusterNames := strings.Split(string(data), "|")
		runOnceOnEachNode(clusterNames)
		return
	})
}

func SequentialTestBeforeSuite(runOnce func()) bool {
	return ginkgo.BeforeSuite(func() {
		DeleteCassandraResourcesInNamespace(Namespace)
		runOnce()
	})
}
