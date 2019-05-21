package adjuster

import (
	"testing"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
	"k8s.io/api/core/v1"

	"encoding/json"
	"fmt"

	"github.com/PaesslerAG/jsonpath"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	containerCPU                      = "$.spec.template.spec.containers[0].resources.requests.cpu"
	containerMemoryRequest            = "$.spec.template.spec.containers[0].resources.requests.memory"
	containerMemoryLimit              = "$.spec.template.spec.containers[0].resources.limits.memory"
	livenessProbeTimeout              = "$.spec.template.spec.containers[0].livenessProbe.timeoutSeconds"
	livenessProbeFailureThreshold     = "$.spec.template.spec.containers[0].livenessProbe.failureThreshold"
	livenessProbeInitialDelaySeconds  = "$.spec.template.spec.containers[0].livenessProbe.initialDelaySeconds"
	livenessProbePeriodSeconds        = "$.spec.template.spec.containers[0].livenessProbe.periodSeconds"
	readinessProbeTimeout             = "$.spec.template.spec.containers[0].readinessProbe.timeoutSeconds"
	readinessProbeFailureThreshold    = "$.spec.template.spec.containers[0].readinessProbe.failureThreshold"
	readinessProbeInitialDelaySeconds = "$.spec.template.spec.containers[0].readinessProbe.initialDelaySeconds"
	readinessProbeSuccessThreshold    = "$.spec.template.spec.containers[0].readinessProbe.successThreshold"
	readinessProbePeriodSeconds       = "$.spec.template.spec.containers[0].readinessProbe.periodSeconds"
	rackReplicas                      = "$.spec.replicas"
	clusterConfigHash                 = "$.spec.template.metadata.annotations.clusterConfigHash"
	bootstrapperImage                 = "$.spec.template.spec.initContainers[0].image"
)

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Operator Suite", test.CreateParallelReporters("operator"))
}

var _ = Describe("cluster events", func() {
	var oldCluster *v1alpha1.Cassandra
	var newCluster *v1alpha1.Cassandra
	var adjuster *Adjuster

	BeforeEach(func() {
		livenessProbe := &v1alpha1.Probe{
			FailureThreshold:    int32(3),
			InitialDelaySeconds: int32(30),
			PeriodSeconds:       int32(30),
			SuccessThreshold:    int32(1),
			TimeoutSeconds:      int32(5),
		}
		readinessProbe := &v1alpha1.Probe{
			FailureThreshold:    int32(3),
			InitialDelaySeconds: int32(30),
			PeriodSeconds:       int32(15),
			SuccessThreshold:    int32(1),
			TimeoutSeconds:      int32(5),
		}
		oldCluster = &v1alpha1.Cassandra{
			Spec: v1alpha1.CassandraSpec{
				Datacenter: ptr.String("ADatacenter"),
				Racks:      []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}},
				Pod: v1alpha1.Pod{
					Image:          ptr.String("anImage"),
					Memory:         resource.MustParse("2Gi"),
					CPU:            resource.MustParse("100m"),
					StorageSize:    resource.MustParse("1Gi"),
					LivenessProbe:  livenessProbe.DeepCopy(),
					ReadinessProbe: readinessProbe.DeepCopy(),
				},
			},
		}
		newCluster = &v1alpha1.Cassandra{
			Spec: v1alpha1.CassandraSpec{
				Datacenter: ptr.String("ADatacenter"),
				Racks:      []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}},
				Pod: v1alpha1.Pod{
					Image:          ptr.String("anImage"),
					Memory:         resource.MustParse("2Gi"),
					CPU:            resource.MustParse("100m"),
					StorageSize:    resource.MustParse("1Gi"),
					LivenessProbe:  livenessProbe.DeepCopy(),
					ReadinessProbe: readinessProbe.DeepCopy(),
				},
			},
		}
		var err error
		adjuster, err = New()
		Expect(err).To(Not(HaveOccurred()))
	})

	Context("a config map hash annotation patch is requested for rack", func() {
		It("should produce an UpdateRack change with the new config map hash", func() {
			rack := v1alpha1.Rack{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}
			configMap := v1.ConfigMap{
				Data: map[string]string{
					"test": "value",
				},
			}
			change := adjuster.CreateConfigMapHashPatchForRack(&rack, &configMap)

			Expect(change.Rack).To(Equal(rack))
			Expect(change.ChangeType).To(Equal(UpdateRack))
			Expect(evaluateJSONPath(clusterConfigHash, change.Patch)).To(Equal("29ab74e6c0e7eb7d55f4d76d92a3f4bab949e0539600ab8f37fdd882fa44cdf4"))
		})
	})

	Context("pod spec change is detected", func() {
		It("should produce a change with updated cpu when cpu specification has changed", func() {
			newCluster.Spec.Pod.CPU = resource.MustParse("110m")
			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[0], UpdateRack, map[string]interface{}{containerCPU: "110m"}, 0))
		})

		It("should produce a change with updated memory when memory specification has changed", func() {
			newCluster.Spec.Pod.Memory = resource.MustParse("1Gi")
			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[0], UpdateRack, map[string]interface{}{containerMemoryRequest: "1Gi", containerMemoryLimit: "1Gi"}, 0))
		})

		It("should produce a change with updated timeout when liveness probe specification has changed", func() {
			newCluster.Spec.Pod.LivenessProbe.FailureThreshold = 5
			newCluster.Spec.Pod.LivenessProbe.InitialDelaySeconds = 99
			newCluster.Spec.Pod.LivenessProbe.PeriodSeconds = 20
			newCluster.Spec.Pod.LivenessProbe.TimeoutSeconds = 10
			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[0],
				UpdateRack,
				map[string]interface{}{
					livenessProbeFailureThreshold:    float64(5),
					livenessProbeInitialDelaySeconds: float64(99),
					livenessProbePeriodSeconds:       float64(20),
					livenessProbeTimeout:             float64(10),
				},
				0))
		})

		It("should produce a change with updated timeout when readiness probe specification has changed", func() {
			newCluster.Spec.Pod.ReadinessProbe.FailureThreshold = 27
			newCluster.Spec.Pod.ReadinessProbe.InitialDelaySeconds = 55
			newCluster.Spec.Pod.ReadinessProbe.PeriodSeconds = 77
			newCluster.Spec.Pod.ReadinessProbe.SuccessThreshold = 80
			newCluster.Spec.Pod.ReadinessProbe.TimeoutSeconds = 4
			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[0],
				UpdateRack, map[string]interface{}{
					readinessProbeFailureThreshold:    float64(27),
					readinessProbeInitialDelaySeconds: float64(55),
					readinessProbePeriodSeconds:       float64(77),
					readinessProbeSuccessThreshold:    float64(80),
					readinessProbeTimeout:             float64(4),
				}, 0))
		})

		It("should produce a patch containing the updated image when the bootstrapper image has been updated", func() {
			newCluster.Spec.Pod.BootstrapperImage = ptr.String("someotherimage")
			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[0], UpdateRack, map[string]interface{}{bootstrapperImage: "someotherimage"}, 0))
		})
	})

	Context("scale-up change is detected", func() {
		Context("single-rack cluster", func() {
			It("should produce a change with the updated number of replicas", func() {
				newCluster.Spec.Racks[0].Replicas = 2
				changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

				Expect(err).ToNot(HaveOccurred())
				Expect(changes).To(HaveLen(1))
				Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[0], UpdateRack, map[string]interface{}{rackReplicas: float64(2)}, 0))
			})
		})

		Context("multiple-rack cluster", func() {
			It("should produce a change for each changed rack with the updated number of replicas", func() {
				oldCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}, {Name: "b", Replicas: 1, StorageClass: "another-storage", Zone: "another-zone"}, {Name: "c", Replicas: 1, StorageClass: "yet-another-storage", Zone: "yet-another-zone"}}
				newCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 2, StorageClass: "some-storage", Zone: "some-zone"}, {Name: "b", Replicas: 1, StorageClass: "another-storage", Zone: "another-zone"}, {Name: "c", Replicas: 3, StorageClass: "yet-another-storage", Zone: "yet-another-zone"}}
				changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

				Expect(err).ToNot(HaveOccurred())
				Expect(changes).To(HaveLen(2))
				Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[0], UpdateRack, map[string]interface{}{rackReplicas: float64(2)}, 0))
				Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[2], UpdateRack, map[string]interface{}{rackReplicas: float64(3)}, 0))
			})
		})
	})

	Context("both pod resource and scale-up changes are detected", func() {
		It("should produce a change for all racks with the updated pod resource and replication", func() {
			oldCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}, {Name: "b", Replicas: 1, StorageClass: "another-storage", Zone: "another-zone"}}
			newCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}, {Name: "b", Replicas: 3, StorageClass: "another-storage", Zone: "another-zone"}}
			newCluster.Spec.Pod.CPU = resource.MustParse("1")
			newCluster.Spec.Pod.Memory = resource.MustParse("999Mi")

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(2))
			Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[0], UpdateRack, map[string]interface{}{rackReplicas: float64(1), containerCPU: "1", containerMemoryRequest: "999Mi", containerMemoryLimit: "999Mi"}, 0))
			Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[1], UpdateRack, map[string]interface{}{rackReplicas: float64(3), containerCPU: "1", containerMemoryRequest: "999Mi", containerMemoryLimit: "999Mi"}, 0))
		})
	})

	Context("nothing has changed in the definition", func() {
		It("should not produce any changes", func() {
			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(0))
		})
	})

	Context("unsupported property change is detected", func() {
		It("should reject the change with an error message when DC is changed", func() {
			newCluster.Spec.Datacenter = ptr.String("other-dc")
			_, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).To(MatchError("changing dc is forbidden. The dc used will continue to be 'ADatacenter'"))
		})

		It("should report that the default DC name will continue to be used when no DC was previously provided", func() {
			oldCluster.Spec.Datacenter = nil
			newCluster.Spec.Datacenter = ptr.String("new-dc")
			_, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).To(MatchError(fmt.Sprintf("changing dc is forbidden. The dc used will continue to be '%s'", v1alpha1.DefaultDCName)))
		})

		It("should reject the change with an error message when Image is changed", func() {
			newCluster.Spec.Pod.Image = ptr.String("other-image")
			_, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).To(MatchError("changing image is forbidden. The image used will continue to be 'anImage'"))
		})

		It("should report that the default image will continue to be used if an image was not previously specified", func() {
			oldCluster.Spec.Pod.Image = nil
			newCluster.Spec.Pod.Image = ptr.String("other-image")

			_, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).To(MatchError(fmt.Sprintf("changing image is forbidden. The image used will continue to be '%s'", v1alpha1.DefaultCassandraImage)))
		})

		It("should reject the change with an error message when UseEmptyDir is changed", func() {
			newCluster.Spec.UseEmptyDir = ptr.Bool(true)
			_, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).To(MatchError("changing useEmptyDir is forbidden. The useEmptyDir used will continue to be 'false'"))
		})

		It("should reject the change with an error message when a rack storageClass is changed", func() {
			newCluster.Spec.Racks[0].StorageClass = "another-storage-class"
			_, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).To(MatchError("changing storageClass for rack 'a' is forbidden. The storageClass used will continue to be 'some-storage'"))
		})

		It("should reject the change with an error message when a rack zone is changed", func() {
			newCluster.Spec.Racks[0].Zone = "another-zone"
			_, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).To(MatchError("changing zone for rack 'a' is forbidden. The zone used will continue to be 'some-zone'"))
		})
	})
	Context("a new rack definition is added", func() {
		It("should produce a change describing the new rack", func() {
			newRack := v1alpha1.Rack{Name: "b", Replicas: 2, Zone: "zone-b", StorageClass: "storage-class-b"}
			newCluster.Spec.Racks = append(newCluster.Spec.Racks, newRack)

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes).To(HaveClusterChange(newRack, AddRack, nil, 0))
		})
	})

	Context("a rack definition is deleted", func() {
		BeforeEach(func() {
			oldCluster.Spec.Racks = []v1alpha1.Rack{
				{Name: "a", Replicas: 1, Zone: "zone-a", StorageClass: "storage-class-a"},
				{Name: "b", Replicas: 1, Zone: "zone-b", StorageClass: "storage-class-b"},
			}
		})

		It("should produce a change describing the rack which was deleted", func() {
			newCluster.Spec.Racks = []v1alpha1.Rack{{Name: "b", Replicas: 1, Zone: "zone-b", StorageClass: "storage-class-b"}}

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes).To(HaveClusterChange(oldCluster.Spec.Racks[0], deleteRack, nil, 0))
		})
	})

	Context("a rack is scaled down", func() {
		It("should produce a change describing the rack which was scaled down, and how many pods should be removed", func() {
			oldCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 2, StorageClass: "some-storage", Zone: "some-zone"}}
			newCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}}

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes).To(HaveClusterChange(newCluster.Spec.Racks[0], scaleDownRack, nil, 1))
		})
	})

	Context("correct ordering of changes", func() {
		It("should perform rack additions before rack deletions", func() {
			oldCluster.Spec.Racks = []v1alpha1.Rack{{Name: "c", Replicas: 1, Zone: "zone-c", StorageClass: "storage-class-c"}}
			newCluster.Spec.Racks = []v1alpha1.Rack{
				{Name: "a", Replicas: 1, Zone: "zone-a", StorageClass: "storage-class-a"},
				{Name: "b", Replicas: 1, Zone: "zone-b", StorageClass: "storage-class-b"},
			}

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(3))
			Expect(changes[0].ChangeType).To(Equal(AddRack))
			Expect(changes[0].Rack.Name).To(Or(Equal("a"), Equal("b")))

			Expect(changes[1].ChangeType).To(Equal(AddRack))
			Expect(changes[1].Rack.Name).To(Or(Equal("a"), Equal("b")))

			Expect(changes[2].ChangeType).To(Equal(deleteRack))
			Expect(changes[2].Rack.Name).To(Equal("c"))
		})

		It("should perform scale down operations before update operations", func() {
			oldCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 2, Zone: "zone-a", StorageClass: "storage-class-a"}}
			newCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 1, Zone: "zone-a", StorageClass: "storage-class-a"}}
			newCluster.Spec.Pod.Memory = resource.MustParse("3Gi")

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(2))
			Expect(changes[0].ChangeType).To(Equal(scaleDownRack))
			Expect(changes[1].ChangeType).To(Equal(UpdateRack))
		})

		It("should perform delete operations before scale down and update operations", func() {
			oldCluster.Spec.Racks = []v1alpha1.Rack{
				{Name: "a", Replicas: 2, Zone: "zone-a", StorageClass: "storage-class-a"},
				{Name: "b", Replicas: 2, Zone: "zone-b", StorageClass: "storage-class-b"},
			}
			newCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 1, Zone: "zone-a", StorageClass: "storage-class-a"}}
			newCluster.Spec.Pod.Memory = resource.MustParse("3Gi")

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(3))
			Expect(changes[0].ChangeType).To(Equal(deleteRack))
			Expect(changes[0].Rack.Name).To(Equal("b"))

			Expect(changes[1].ChangeType).To(Equal(scaleDownRack))
			Expect(changes[1].Rack.Name).To(Equal("a"))

			Expect(changes[2].ChangeType).To(Equal(UpdateRack))
			Expect(changes[2].Rack.Name).To(Equal("a"))
		})

		It("should perform add operations before scale down and update operations", func() {
			oldCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 2, Zone: "zone-a", StorageClass: "storage-class-a"}, {Name: "b", Replicas: 2, Zone: "zone-b", StorageClass: "storage-class-b"}}
			newCluster.Spec.Racks = []v1alpha1.Rack{
				{Name: "a", Replicas: 1, Zone: "zone-a", StorageClass: "storage-class-a"},
				{Name: "b", Replicas: 1, Zone: "zone-b", StorageClass: "storage-class-b"},
				{Name: "c", Replicas: 1, Zone: "zone-c", StorageClass: "storage-class-c"},
			}
			newCluster.Spec.Pod.Memory = resource.MustParse("3Gi")

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(5))
			Expect(changes[0].ChangeType).To(Equal(AddRack))
			Expect(changes[0].Rack.Name).To(Equal("c"))

			Expect(changes[1].ChangeType).To(Equal(scaleDownRack))
			Expect(changes[1].Rack.Name).To(Or(Equal("a"), Equal("b")))

			Expect(changes[2].ChangeType).To(Equal(scaleDownRack))
			Expect(changes[2].Rack.Name).To(Or(Equal("a"), Equal("b")))

			Expect(changes[3].ChangeType).To(Equal(UpdateRack))
			Expect(changes[3].Rack.Name).To(Or(Equal("a"), Equal("b")))

			Expect(changes[4].ChangeType).To(Equal(UpdateRack))
			Expect(changes[4].Rack.Name).To(Or(Equal("a"), Equal("b")))
		})

		It("should perform adds before delete, scale down and update operations", func() {
			oldCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 2, Zone: "zone-a", StorageClass: "storage-class-a"}, {Name: "b", Replicas: 2, Zone: "zone-b", StorageClass: "storage-class-b"}}
			newCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 1, Zone: "zone-a", StorageClass: "storage-class-a"}, {Name: "c", Replicas: 1, Zone: "zone-c", StorageClass: "storage-class-c"}}
			newCluster.Spec.Pod.Memory = resource.MustParse("3Gi")

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(4))
			Expect(changes[0].ChangeType).To(Equal(AddRack))
			Expect(changes[0].Rack.Name).To(Equal("c"))

			Expect(changes[1].ChangeType).To(Equal(deleteRack))
			Expect(changes[1].Rack.Name).To(Equal("b"))

			Expect(changes[2].ChangeType).To(Equal(scaleDownRack))
			Expect(changes[2].Rack.Name).To(Equal("a"))

			Expect(changes[3].ChangeType).To(Equal(UpdateRack))
			Expect(changes[3].Rack.Name).To(Equal("a"))
		})

		It("should treat a scale up and update as a single update operation", func() {
			oldCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 1, Zone: "zone-a", StorageClass: "storage-class-a"}}
			newCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 2, Zone: "zone-a", StorageClass: "storage-class-a"}}
			newCluster.Spec.Pod.Memory = resource.MustParse("3Gi")

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes[0].ChangeType).To(Equal(UpdateRack))
			Expect(changes[0].Rack.Name).To(Equal("a"))
		})

		It("should order racks changes in the order in which the racks were defined in the old cluster state", func() {
			oldCluster.Spec.Racks = []v1alpha1.Rack{{Name: "a", Replicas: 1, Zone: "zone-a", StorageClass: "storage-class-a"}, {Name: "b", Replicas: 1, Zone: "zone-b", StorageClass: "storage-class-b"}, {Name: "c", Replicas: 1, Zone: "zone-c", StorageClass: "storage-class-c"}}
			newCluster.Spec.Racks = []v1alpha1.Rack{{Name: "c", Replicas: 1, Zone: "zone-c", StorageClass: "storage-class-c"}, {Name: "a", Replicas: 1, Zone: "zone-a", StorageClass: "storage-class-a"}, {Name: "b", Replicas: 1, Zone: "zone-b", StorageClass: "storage-class-b"}}
			newCluster.Spec.Pod.CPU = resource.MustParse("101m")

			changes, err := adjuster.ChangesForCluster(oldCluster, newCluster)

			Expect(err).ToNot(HaveOccurred())
			Expect(changes).To(HaveLen(3))
			Expect(changes[0].Rack.Name).To(Equal("a"))
			Expect(changes[1].Rack.Name).To(Equal("b"))
			Expect(changes[2].Rack.Name).To(Equal("c"))
		})
	})
})

//
// HaveClusterChange matcher
//
func HaveClusterChange(rack v1alpha1.Rack, changeType ClusterChangeType, patchExpectations map[string]interface{}, nodesToScaleDown int) types.GomegaMatcher {
	return &haveClusterChange{rack, changeType, patchExpectations, nodesToScaleDown}
}

type haveClusterChange struct {
	rack              v1alpha1.Rack
	changeType        ClusterChangeType
	patchExpectations map[string]interface{}
	nodesToScaleDown  int
}

func (matcher *haveClusterChange) Match(actual interface{}) (success bool, err error) {
	changes := actual.([]ClusterChange)
	if change := matcher.findClusterChange(&matcher.rack, changes); change != nil {
		if change.Rack != matcher.rack {
			return false, fmt.Errorf("expected rack %v to match, but found %v", matcher.rack, change.Rack)
		}

		if change.ChangeType != matcher.changeType {
			return false, fmt.Errorf("expected change type %s, but found %s", matcher.changeType, change.ChangeType)
		}

		if change.nodesToScaleDown != matcher.nodesToScaleDown {
			return false, fmt.Errorf("expected to scale down %d nodes, but found %d", matcher.nodesToScaleDown, change.nodesToScaleDown)
		}

		for jsonPath, expectedValue := range matcher.patchExpectations {
			foundValue, err := evaluateJSONPath(jsonPath, change.Patch)
			if err != nil {
				return false, err
			}

			if foundValue != expectedValue {
				return false, fmt.Errorf("expected value %v at json path %s, but got %v. Change patch: %v", expectedValue, jsonPath, foundValue, change.Patch)
			}
		}
	} else {
		return false, fmt.Errorf("no matching change for rack %s in changes %v", matcher.rack.Name, changes)
	}
	return true, nil
}

func (matcher *haveClusterChange) findClusterChange(rackToFind *v1alpha1.Rack, changes []ClusterChange) *ClusterChange {
	for _, change := range changes {
		if change.Rack.Name == matcher.rack.Name {
			return &change
		}
	}
	return nil
}

func (matcher *haveClusterChange) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Actual clusterChange: %s", actual)
}

func (matcher *haveClusterChange) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Actual clusterChange: %s", actual)
}

func evaluateJSONPath(path, document string) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(document), &v); err != nil {
		return nil, err
	}

	return jsonpath.Get(path, v)
}
