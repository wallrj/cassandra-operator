package parallel_modification

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
	. "github.com/sky-uk/cassandra-operator/cassandra-operator/test/e2e"
	coreV1 "k8s.io/api/core/v1"
	"testing"
	"time"
)

var (
	cluster1      *TestCluster
	cluster2      *TestCluster
	cluster3      *TestCluster
	cluster4      *TestCluster
	cluster5      *TestCluster
	cluster6      *TestCluster
	testStartTime time.Time
)

func TestParallelModification(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "E2E Suite (Parallel Modification Tests)", test.CreateSequentialReporters("e2e_parallel_modification"))
}

func defineClusters(cluster1Name, cluster2Name, cluster3Name, cluster4Name, cluster5Name, cluster6Name string) (cluster1, cluster2, cluster3, cluster4, cluster5, cluster6 *TestCluster) {
	cluster1 = &TestCluster{
		Name:  cluster1Name,
		Racks: []v1alpha1.Rack{Rack("a", 1)},
	}
	cluster2 = &TestCluster{
		Name:  cluster2Name,
		Racks: []v1alpha1.Rack{Rack("a", 1)},
	}
	cluster3 = &TestCluster{
		Name:  cluster3Name,
		Racks: []v1alpha1.Rack{Rack("a", 1)},
	}
	cluster4 = &TestCluster{
		Name:  cluster4Name,
		Racks: []v1alpha1.Rack{Rack("a", 1)},
	}
	cluster5 = &TestCluster{
		Name:  cluster5Name,
		Racks: []v1alpha1.Rack{Rack("a", 1)},
	}
	cluster6 = &TestCluster{
		Name:  cluster6Name,
		Racks: []v1alpha1.Rack{Rack("a", 1)},
	}
	return
}

func createClusters(clusters ...*TestCluster) {
	// create the clusters in parallel
	for _, clusterToCreate := range clusters {
		AClusterWithName(clusterToCreate.Name).AndRacks(clusterToCreate.Racks).UsingEmptyDir().IsDefined()
	}

	for _, clusterDefined := range clusters {
		EventuallyClusterIsCreatedWithRacks(Namespace, clusterDefined.Name, clusterDefined.Racks)
	}
}

var _ = SequentialTestBeforeSuite(func() {
	cluster1, cluster2, cluster3, cluster4, cluster5, cluster6  = defineClusters(AClusterName(), AClusterName(), AClusterName(), AClusterName(), AClusterName(), AClusterName())
	createClusters(cluster1, cluster2, cluster3, cluster4, cluster5, cluster6 )
})

var _ = Context("cluster modifications in parallel", func() {

	BeforeEach(func() {
		testStartTime = time.Now()
	})

	JustAfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			PrintDiagnosis(Namespace, testStartTime, cluster1.Name)
		}
	})

	It("should record all modification event attempts", func() {
		modificationTime := time.Now()
		TheImageImmutablePropertyIsChangedTo(Namespace, cluster1.Name, "another-image")

		//time.Sleep(time.Minute)
		modification2Time := time.Now()
		TheDcImmutablePropertyIsChangedTo(Namespace, cluster2.Name, "another-dc")

		modification3Time := time.Now()
		TheImageImmutablePropertyIsChangedTo(Namespace, cluster3.Name, "yet-another-image")

		modification4Time := time.Now()
		TheImageImmutablePropertyIsChangedTo(Namespace, cluster4.Name, "another-image")

		modification5Time := time.Now()
		TheDcImmutablePropertyIsChangedTo(Namespace, cluster5.Name, "another-dc")

		modification6Time := time.Now()
		TheImageImmutablePropertyIsChangedTo(Namespace, cluster6.Name, "yet-another-image")

		// then
		By("recording a warning event about the forbidden change for all attempted change")
		Eventually(CassandraEventsFor(Namespace, cluster1.Name), 30*time.Second, CheckInterval).Should(HaveEvent(EventExpectation{
			Type:                 coreV1.EventTypeWarning,
			Reason:               cluster.InvalidChangeEvent,
			Message:              "changing image is forbidden",
			LastTimestampCloseTo: modificationTime,
		}))
		Eventually(CassandraEventsFor(Namespace, cluster2.Name), 30*time.Second, CheckInterval).Should(HaveEvent(EventExpectation{
			Type:                 coreV1.EventTypeWarning,
			Reason:               cluster.InvalidChangeEvent,
			Message:              "changing dc is forbidden",
			LastTimestampCloseTo: modification2Time,
		}))
		Eventually(CassandraEventsFor(Namespace, cluster3.Name), 30*time.Second, CheckInterval).Should(HaveEvent(EventExpectation{
			Type:                 coreV1.EventTypeWarning,
			Reason:               cluster.InvalidChangeEvent,
			Message:              "changing image is forbidden",
			LastTimestampCloseTo: modification3Time,
		}))
		Eventually(CassandraEventsFor(Namespace, cluster4.Name), 30*time.Second, CheckInterval).Should(HaveEvent(EventExpectation{
			Type:                 coreV1.EventTypeWarning,
			Reason:               cluster.InvalidChangeEvent,
			Message:              "changing image is forbidden",
			LastTimestampCloseTo: modification4Time,
		}))
		Eventually(CassandraEventsFor(Namespace, cluster5.Name), 30*time.Second, CheckInterval).Should(HaveEvent(EventExpectation{
			Type:                 coreV1.EventTypeWarning,
			Reason:               cluster.InvalidChangeEvent,
			Message:              "changing dc is forbidden",
			LastTimestampCloseTo: modification5Time,
		}))
		Eventually(CassandraEventsFor(Namespace, cluster6.Name), 30*time.Second, CheckInterval).Should(HaveEvent(EventExpectation{
			Type:                 coreV1.EventTypeWarning,
			Reason:               cluster.InvalidChangeEvent,
			Message:              "changing image is forbidden",
			LastTimestampCloseTo: modification6Time,
		}))
	})

	It("should record all modification events", func() {
		clusters := []*TestCluster{cluster1, cluster2, cluster3, cluster4, cluster5, cluster6}
		modificationTime := time.Now()
		for _, clusterUnderTest := range clusters {
			TheCustomConfigIsDeletedForCluster(Namespace, clusterUnderTest.Name)
		}

		// then
		By("recording a warning event about the forbidden change for all attempted change")
		for _, clusterUnderTest := range clusters {
			Eventually(CassandraEventsFor(Namespace, clusterUnderTest.Name), 30*time.Second, CheckInterval).Should(HaveEvent(EventExpectation{
				Type:                 coreV1.EventTypeNormal,
				Reason:               cluster.ClusterUpdateEvent,
				Message:              fmt.Sprintf("Custom config deleted for cluster %s.%s", Namespace, clusterUnderTest.Name),
				LastTimestampCloseTo: modificationTime,
			}))
		}
	})

	FIt("should record both valid and modification events", func() {
		clusters := []*TestCluster{cluster1, cluster2, cluster3, cluster4, cluster5, cluster6}
		modificationTime := time.Now()
		for _, clusterUnderTest := range clusters {
			TheImageImmutablePropertyIsChangedTo(Namespace, clusterUnderTest.Name, "another-image")
			TheCustomConfigIsDeletedForCluster(Namespace, clusterUnderTest.Name)
		}

		// then
		By("recording a warning event about the forbidden change for all attempted change")
		for _, clusterUnderTest := range clusters {
			Eventually(CassandraEventsFor(Namespace, clusterUnderTest.Name), 30*time.Second, CheckInterval).Should(HaveEvent(EventExpectation{
				Type:                 coreV1.EventTypeNormal,
				Reason:               cluster.ClusterUpdateEvent,
				Message:              fmt.Sprintf("Custom config deleted for cluster %s.%s", Namespace, clusterUnderTest.Name),
				LastTimestampCloseTo: modificationTime,
			}))
			Eventually(CassandraEventsFor(Namespace, clusterUnderTest.Name), 30*time.Second, CheckInterval).Should(HaveEvent(EventExpectation{
				Type:                 coreV1.EventTypeWarning,
				Reason:               cluster.InvalidChangeEvent,
				Message:              "changing image is forbidden",
				LastTimestampCloseTo: modificationTime,
			}))
		}
	})

})
