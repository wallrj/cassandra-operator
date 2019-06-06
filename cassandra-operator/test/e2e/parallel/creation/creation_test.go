package creation

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
	"k8s.io/apimachinery/pkg/api/resource"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
	. "github.com/sky-uk/cassandra-operator/cassandra-operator/test/e2e"
)

var (
	multipleRacksCluster *TestCluster
	emptyDirCluster      *TestCluster
	testStartTime        time.Time
)

func TestCreation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "E2E Suite (Creation Tests)", test.CreateParallelReporters("e2e_creation"))
}

func defineClusters(multipleRacksClusterName, emptyDirClusterName string) (multipleRacksCluster, emptyDirCluster *TestCluster) {
	multipleRacksCluster = &TestCluster{
		Name:                multipleRacksClusterName,
		Racks:               []v1alpha1.Rack{Rack("a", 2), Rack("b", 1)},
		ExtraConfigFileName: "extraConfigFile",
	}
	emptyDirCluster = &TestCluster{
		Name:  emptyDirClusterName,
		Racks: []v1alpha1.Rack{Rack("a", 1)},
	}
	return
}

func createClustersInParallel(multipleRacksCluster, emptyDirCluster *TestCluster) {
	extraFile := &ExtraConfigFile{Name: multipleRacksCluster.ExtraConfigFileName, Content: "some content"}

	AClusterWithName(multipleRacksCluster.Name).AndClusterSpec(&v1alpha1.CassandraSpec{
		Datacenter: ptr.String("custom-dc"),
		Pod: v1alpha1.Pod{
			BootstrapperImage: &CassandraBootstrapperImageName,
			SidecarImage:      &CassandraSidecarImageName,
			Image:             &CassandraImageName,
			Memory:            resource.MustParse("987Mi"),
			CPU:               resource.MustParse("1m"),
			StorageSize:       resource.MustParse("100Mi"),
			LivenessProbe: &v1alpha1.Probe{
				FailureThreshold:    CassandraLivenessProbeFailureThreshold,
				TimeoutSeconds:      7,
				InitialDelaySeconds: CassandraInitialDelay,
				PeriodSeconds:       CassandraLivenessPeriod,
			},
			ReadinessProbe: &v1alpha1.Probe{
				FailureThreshold:    CassandraReadinessProbeFailureThreshold,
				TimeoutSeconds:      6,
				InitialDelaySeconds: CassandraInitialDelay,
				PeriodSeconds:       CassandraReadinessPeriod,
			},
		},
	}).AndRacks(multipleRacksCluster.Racks).AndCustomConfig(extraFile).IsDefined()

	AClusterWithName(emptyDirCluster.Name).AndRacks(emptyDirCluster.Racks).UsingEmptyDir().WithoutCustomConfig().IsDefined()
}

var _ = ParallelTestBeforeSuite(func() []TestCluster {
	multipleRacksCluster, emptyDirCluster = defineClusters(AClusterName(), AClusterName())
	createClustersInParallel(multipleRacksCluster, emptyDirCluster)
	return []TestCluster{*multipleRacksCluster, *emptyDirCluster}
}, func(clusterNames []string) {
	multipleRacksCluster, emptyDirCluster = defineClusters(clusterNames[0], clusterNames[1])
})

var _ = Context("When a cluster with a given name doesn't already exist", func() {

	BeforeEach(func() {
		testStartTime = time.Now()
		Eventually(PodReadyForCluster(Namespace, multipleRacksCluster.Name), 3*NodeStartDuration, CheckInterval).
			Should(Equal(3), fmt.Sprintf("For cluster %s", multipleRacksCluster.Name))
		Eventually(PodReadyForCluster(Namespace, emptyDirCluster.Name), NodeStartDuration, CheckInterval).
			Should(Equal(1), fmt.Sprintf("For cluster %s", emptyDirCluster.Name))
	})

	JustAfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			PrintDiagnosis(Namespace, testStartTime, multipleRacksCluster.Name, emptyDirCluster.Name)
		}
	})

	It("should create all the required Kubernetes resources according to the cassandra specs", func() {
		// then
		By("creating all pods only once")
		Expect(PodRestartForCluster(Namespace, multipleRacksCluster.Name)()).Should(Equal(0))

		By("creating pods with the specified resources")
		Expect(PodsForCluster(Namespace, multipleRacksCluster.Name)()).Should(Each(And(
			HaveContainer(ContainerExpectation{
				ImageName:                      CassandraImageName,
				ContainerName:                  "cassandra",
				MemoryRequest:                  "987Mi",
				MemoryLimit:                    "987Mi",
				CPURequest:                     "1m",
				LivenessProbePeriod:            DurationSeconds(CassandraLivenessPeriod),
				LivenessProbeFailureThreshold:  CassandraLivenessProbeFailureThreshold,
				LivenessProbeInitialDelay:      DurationSeconds(CassandraInitialDelay),
				LivenessProbeTimeout:           7 * time.Second,
				ReadinessProbeTimeout:          6 * time.Second,
				ReadinessProbePeriod:           DurationSeconds(CassandraReadinessPeriod),
				ReadinessProbeFailureThreshold: CassandraReadinessProbeFailureThreshold,
				ReadinessProbeInitialDelay:     DurationSeconds(CassandraInitialDelay),
				ReadinessProbeSuccessThreshold: 1,
				ContainerPorts:                 map[string]int{"internode": 7000, "jmx-exporter": 7070, "cassandra-jmx": 7199, "jolokia": 7777, "client": 9042}})),
		))

		By("creating a StatefulSet for each rack")
		Expect(StatefulSetsForCluster(Namespace, multipleRacksCluster.Name)()).Should(Each(And(
			BeCreatedWithServiceName(multipleRacksCluster.Name),
			HaveLabel("sky.uk/cassandra-operator", multipleRacksCluster.Name),
		)))

		By("creating a headless service for the cluster")
		Expect(HeadlessServiceForCluster(Namespace, multipleRacksCluster.Name)()).Should(And(
			Not(BeNil()),
			HaveLabel("sky.uk/cassandra-operator", multipleRacksCluster.Name)),
		)

		By("creating a persistent volume claim for each StatefulSet with the requested storage capacity")
		Expect(PersistentVolumeClaimsForCluster(Namespace, multipleRacksCluster.Name)()).Should(And(
			HaveLen(3),
			Each(HaveLabel("sky.uk/cassandra-operator", multipleRacksCluster.Name)),
			Each(HaveStorageCapacity("100Mi"))))

		if !UseMockedImage {
			By("creating a cluster with the specified datacenter")
			Eventually(DataCenterForCluster(Namespace, multipleRacksCluster.Name), NodeStartDuration, CheckInterval).Should(Equal("custom-dc"))
		}
	})

	It("should copy custom config files into a cassandra config directory within pods", func() {
		Expect(FileExistsInConfigurationDirectory(Namespace, PodName(multipleRacksCluster.Name, "a", 0), filepath.Base(multipleRacksCluster.ExtraConfigFileName))()).To(BeTrue())
	})

	It("should spread out the nodes in different locations", func() {
		By("creating as many racks as locations")
		By("creating the same number of nodes for each rack")
		Expect(RacksForCluster(Namespace, multipleRacksCluster.Name)()).Should(And(
			HaveLen(2),
			HaveKeyWithValue("a", []string{PodName(multipleRacksCluster.Name, "a", 0), PodName(multipleRacksCluster.Name, "a", 1)}),
			HaveKeyWithValue("b", []string{PodName(multipleRacksCluster.Name, "b", 0)})))
	})

	It("should create the pods on different nodes", func() {
		Expect(UniqueNodesUsed(Namespace, multipleRacksCluster.Name)).Should(HaveLen(3))
	})

	It("should report metrics for all clusters", func() {
		By("exposing node and rack information in metrics")
		Eventually(OperatorMetrics(Namespace), 60*time.Second, CheckInterval).Should(ReportAClusterWith([]MetricAssertion{
			ClusterSizeMetric(Namespace, multipleRacksCluster.Name, 3),
			LiveAndNormalNodeMetric(Namespace, multipleRacksCluster.Name, PodName(multipleRacksCluster.Name, "a", 0), "a", 1),
			LiveAndNormalNodeMetric(Namespace, multipleRacksCluster.Name, PodName(multipleRacksCluster.Name, "a", 1), "a", 1),
			LiveAndNormalNodeMetric(Namespace, multipleRacksCluster.Name, PodName(multipleRacksCluster.Name, "b", 0), "b", 1),
		}))

		By("exposing metrics for the other cluster")
		Eventually(OperatorMetrics(Namespace), 60*time.Second, CheckInterval).Should(ReportAClusterWith([]MetricAssertion{
			ClusterSizeMetric(Namespace, emptyDirCluster.Name, 1),
			LiveAndNormalNodeMetric(Namespace, emptyDirCluster.Name, PodName(emptyDirCluster.Name, "a", 0), "a", 1),
		}))
	})

	It("should not create persistent volume claims for an emptydir cluster", func() {
		Expect(PersistentVolumeClaimsForCluster(Namespace, emptyDirCluster.Name)()).Should(BeEmpty())
	})

	It("should add an annotation named customConfigHash for clusters with custom config", func() {
		Expect(PodsForCluster(Namespace, multipleRacksCluster.Name)()).Should(Each(
			HaveAnnotation("clusterConfigHash"),
		))
	})

	It("should not add an annotation named customConfigHash for clusters without custom config", func() {
		Expect(PodsForCluster(Namespace, emptyDirCluster.Name)()).Should(Each(
			Not(HaveAnnotation("clusterConfigHash")),
		))
	})
})
