package cluster

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	v1alpha1helpers "github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1/helpers"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
)

const (
	CLUSTER   = "mycluster"
	NAMESPACE = "mynamespace"
)

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Cluster Suite", test.CreateParallelReporters("cluster"))
}

var _ = Describe("cluster construction", func() {
	var clusterDef *v1alpha1.Cassandra
	BeforeEach(func() {
		retentionPeriod := int32(7)
		clusterDef = &v1alpha1.Cassandra{
			ObjectMeta: metaV1.ObjectMeta{Name: CLUSTER, Namespace: NAMESPACE},
			Spec: v1alpha1.CassandraSpec{
				Racks: []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}, {Name: "b", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}},
				Pod: v1alpha1.Pod{
					Memory:      resource.MustParse("1Gi"),
					CPU:         resource.MustParse("100m"),
					StorageSize: resource.MustParse("1Gi"),
				},
				Snapshot: &v1alpha1.Snapshot{
					Schedule:  "1 23 * * *",
					Keyspaces: []string{"k1"},
					RetentionPolicy: &v1alpha1.RetentionPolicy{
						Enabled:             ptr.Bool(true),
						RetentionPeriodDays: &retentionPeriod,
					},
				},
			},
		}
	})

	Context("config validation", func() {
		It("should reject a configuration with a rack with zero replicas", func() {
			clusterDef.Spec.Racks = []v1alpha1.Rack{{Name: "a"}}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid rack replicas value 0 provided for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a configuration with no pod memory property", func() {
			clusterDef.Spec.Pod.Memory = resource.Quantity{}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("no podMemory property provided for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should allow a configuration with no pod CPU property", func() {
			clusterDef.Spec.Pod.CPU = resource.Quantity{}
			_, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should use the 3.11 version of the apache cassandra image if one is not supplied for the cluster", func() {
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(*cluster.definition.Spec.Pod.Image).To(Equal("cassandra:3.11"))
		})

		It("should use the specified version of the cassandra image if one is given", func() {
			clusterDef.Spec.Pod.Image = ptr.String("somerepo/someimage:v1.0")
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(*cluster.definition.Spec.Pod.Image).To(Equal("somerepo/someimage:v1.0"))
		})

		It("should use the latest version of the cassandra bootstrapper image if one is not supplied for the cluster", func() {
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(*cluster.definition.Spec.Pod.BootstrapperImage).To(Equal("skyuk/cassandra-bootstrapper:latest"))
		})

		It("should use the specified version of the cassandra bootstrapper image if one is given", func() {
			clusterDef.Spec.Pod.BootstrapperImage = ptr.String("somerepo/some-bootstrapper-image:v1.0")
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(*cluster.definition.Spec.Pod.BootstrapperImage).To(Equal("somerepo/some-bootstrapper-image:v1.0"))
		})

		It("should set the default liveness probe values if it is not configured for the cluster", func() {
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(cluster.definition.Spec.Pod.LivenessProbe.FailureThreshold).To(Equal(ptr.Int32(3)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.InitialDelaySeconds).To(Equal(ptr.Int32(30)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.PeriodSeconds).To(Equal(ptr.Int32(30)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.SuccessThreshold).To(Equal(ptr.Int32(1)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.TimeoutSeconds).To(Equal(ptr.Int32(5)))
		})

		It("should set the default readiness probe values if it is not configured for the cluster", func() {
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.FailureThreshold).To(Equal(ptr.Int32(3)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.InitialDelaySeconds).To(Equal(ptr.Int32(30)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.PeriodSeconds).To(Equal(ptr.Int32(15)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.SuccessThreshold).To(Equal(ptr.Int32(1)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.TimeoutSeconds).To(Equal(ptr.Int32(5)))
		})

		It("should set the default liveness probe values if the liveness probe is present but has unspecified values", func() {
			clusterDef.Spec.Pod.LivenessProbe = &v1alpha1.Probe{}
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(cluster.definition.Spec.Pod.LivenessProbe.FailureThreshold).To(Equal(ptr.Int32(3)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.InitialDelaySeconds).To(Equal(ptr.Int32(30)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.PeriodSeconds).To(Equal(ptr.Int32(30)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.SuccessThreshold).To(Equal(ptr.Int32(1)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.TimeoutSeconds).To(Equal(ptr.Int32(5)))
		})

		It("should set the default readiness probe values if the readiness probe is present but has unspecified values", func() {
			clusterDef.Spec.Pod.ReadinessProbe = &v1alpha1.Probe{}
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.FailureThreshold).To(Equal(ptr.Int32(3)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.InitialDelaySeconds).To(Equal(ptr.Int32(30)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.PeriodSeconds).To(Equal(ptr.Int32(15)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.SuccessThreshold).To(Equal(ptr.Int32(1)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.TimeoutSeconds).To(Equal(ptr.Int32(5)))
		})

		It("should use the specified liveness probe values if they are given", func() {
			clusterDef.Spec.Pod.LivenessProbe = &v1alpha1.Probe{
				SuccessThreshold:    ptr.Int32(1),
				PeriodSeconds:       ptr.Int32(2),
				InitialDelaySeconds: ptr.Int32(3),
				FailureThreshold:    ptr.Int32(4),
				TimeoutSeconds:      ptr.Int32(5),
			}
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(cluster.definition.Spec.Pod.LivenessProbe.SuccessThreshold).To(Equal(ptr.Int32(1)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.PeriodSeconds).To(Equal(ptr.Int32(2)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.InitialDelaySeconds).To(Equal(ptr.Int32(3)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.FailureThreshold).To(Equal(ptr.Int32(4)))
			Expect(cluster.definition.Spec.Pod.LivenessProbe.TimeoutSeconds).To(Equal(ptr.Int32(5)))
		})

		It("should use the specified readiness probe values if they are given", func() {
			clusterDef.Spec.Pod.ReadinessProbe = &v1alpha1.Probe{
				SuccessThreshold:    ptr.Int32(1),
				PeriodSeconds:       ptr.Int32(2),
				InitialDelaySeconds: ptr.Int32(3),
				FailureThreshold:    ptr.Int32(4),
				TimeoutSeconds:      ptr.Int32(5),
			}
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.SuccessThreshold).To(Equal(ptr.Int32(1)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.PeriodSeconds).To(Equal(ptr.Int32(2)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.InitialDelaySeconds).To(Equal(ptr.Int32(3)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.FailureThreshold).To(Equal(ptr.Int32(4)))
			Expect(cluster.definition.Spec.Pod.ReadinessProbe.TimeoutSeconds).To(Equal(ptr.Int32(5)))
		})

		It("should reject a liveness probe which does not have a success threshold of 1", func() {
			clusterDef.Spec.Pod.LivenessProbe = &v1alpha1.Probe{
				SuccessThreshold: ptr.Int32(3),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid success threshold for liveness probe, must be set to 1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a liveness probe which has a negative success threshold", func() {
			clusterDef.Spec.Pod.LivenessProbe = &v1alpha1.Probe{
				SuccessThreshold: ptr.Int32(-1),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid success threshold for liveness probe, must be set to 1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a liveness probe which has a failure threshold less than 1", func() {
			clusterDef.Spec.Pod.LivenessProbe = &v1alpha1.Probe{
				FailureThreshold: ptr.Int32(-1),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid failure threshold for liveness probe, must be 1 or greater, got -1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a liveness probe which has an inital delay seconds less than 1", func() {
			clusterDef.Spec.Pod.LivenessProbe = &v1alpha1.Probe{
				InitialDelaySeconds: ptr.Int32(-1),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid initial delay for liveness probe, must be 1 or greater, got -1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a liveness probe which has a period seconds less than 1", func() {
			clusterDef.Spec.Pod.LivenessProbe = &v1alpha1.Probe{
				PeriodSeconds: ptr.Int32(-1),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid period seconds for liveness probe, must be 1 or greater, got -1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a liveness probe which has a timeout seconds less than 1", func() {
			clusterDef.Spec.Pod.LivenessProbe = &v1alpha1.Probe{
				TimeoutSeconds: ptr.Int32(-1),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid timeout seconds for liveness probe, must be 1 or greater, got -1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a readiness probe which has a negative success threshold", func() {
			clusterDef.Spec.Pod.ReadinessProbe = &v1alpha1.Probe{
				SuccessThreshold: ptr.Int32(-1),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid success threshold for readiness probe, must be 1 or greater, got -1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a readiness probe which has a failure threshold less than 1", func() {
			clusterDef.Spec.Pod.ReadinessProbe = &v1alpha1.Probe{
				FailureThreshold: ptr.Int32(-1),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid failure threshold for readiness probe, must be 1 or greater, got -1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a readiness probe which has an inital delay seconds less than 1", func() {
			clusterDef.Spec.Pod.ReadinessProbe = &v1alpha1.Probe{
				InitialDelaySeconds: ptr.Int32(-1),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid initial delay for readiness probe, must be 1 or greater, got -1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a readiness probe which has a period seconds less than 1", func() {
			clusterDef.Spec.Pod.ReadinessProbe = &v1alpha1.Probe{
				PeriodSeconds: ptr.Int32(-1),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid period seconds for readiness probe, must be 1 or greater, got -1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a readiness probe which has a timeout seconds less than 1", func() {
			clusterDef.Spec.Pod.ReadinessProbe = &v1alpha1.Probe{
				TimeoutSeconds: ptr.Int32(-1),
			}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("invalid timeout seconds for readiness probe, must be 1 or greater, got -1 for Cassandra cluster definition: mynamespace.mycluster"))
		})

		It("should reject a configuration where no racks are provided", func() {
			clusterDef.Spec.Racks = []v1alpha1.Rack{}
			_, err := ACluster(clusterDef)
			Expect(err).To(MatchError("no racks specified for cluster: mynamespace.mycluster"))
		})

		Context("useEmptyDir is true", func() {
			BeforeEach(func() {
				clusterDef.Spec.UseEmptyDir = ptr.Bool(true)
			})

			It("should accept a configuration with no pod storage", func() {
				clusterDef.Spec.Pod.StorageSize = resource.MustParse("0")
				_, err := ACluster(clusterDef)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should reject a configuration where podStorageSize is present", func() {
				clusterDef.Spec.Pod.StorageSize = resource.MustParse("10Gi")
				_, err := ACluster(clusterDef)
				Expect(err).To(MatchError("podStorageSize property provided when useEmptyDir is true for Cassandra cluster definition: mynamespace.mycluster"))
			})

			It("should respect the useEmptyDir flag if the operator is configured to allow emptyDir and podStorageSize is not set", func() {
				clusterDef.Spec.Pod.StorageSize = resource.Quantity{}
				cluster, err := ACluster(clusterDef)
				Expect(err).ToNot(HaveOccurred())
				Expect(v1alpha1helpers.UseEmptyDir(cluster.definition)).To(BeTrue())
			})
		})

		Context("useEmptyDir is false", func() {
			BeforeEach(func() {
				clusterDef.Spec.UseEmptyDir = ptr.Bool(false)
			})

			It("should reject a configuration with no pod storage size property", func() {
				clusterDef.Spec.Pod.StorageSize = resource.Quantity{}
				_, err := ACluster(clusterDef)
				Expect(err).To(MatchError("no podStorageSize property provided and useEmptyDir false for Cassandra cluster definition: mynamespace.mycluster"))
			})

			It("should reject a configuration with racks having no storageclass", func() {
				clusterDef.Spec.Racks = []v1alpha1.Rack{{Name: "a", StorageClass: "", Zone: "a", Replicas: 1}}
				_, err := ACluster(clusterDef)
				Expect(err).To(MatchError("rack named 'a' with no storage class specified, either set useEmptyDir to true or specify storage class: mynamespace.mycluster"))
			})

			It("should reject a configuration with racks having no zone", func() {
				clusterDef.Spec.Racks = []v1alpha1.Rack{{Name: "a", StorageClass: "some-storage-class", Zone: "", Replicas: 1}}
				_, err := ACluster(clusterDef)
				Expect(err).To(MatchError("rack named 'a' with no zone specified, either set useEmptyDir to true or specify zone: mynamespace.mycluster"))
			})
		})

		Context("snapshot config", func() {
			It("should be rejected when schedule is not provided", func() {
				clusterDef.Spec.Snapshot.Schedule = ""
				_, err := ACluster(clusterDef)
				Expect(err).To(MatchError("no snapshot schedule property provided for Cassandra cluster definition: mynamespace.mycluster"))
			})

			It("should be rejected when schedule is not a cron expression", func() {
				clusterDef.Spec.Snapshot.Schedule = "1 2 3 x"
				_, err := ACluster(clusterDef)
				Expect(err).To(MatchError("invalid snapshot schedule, must be a cron expression but got '1 2 3 x' for Cassandra cluster definition: mynamespace.mycluster"))
			})

			It("should be rejected when cleanup schedule is not a cron expression", func() {
				clusterDef.Spec.Snapshot.RetentionPolicy.CleanupSchedule = "1 x y z"
				_, err := ACluster(clusterDef)
				Expect(err).To(MatchError("invalid snapshot cleanup schedule, must be a cron expression but got '1 x y z' for Cassandra cluster definition: mynamespace.mycluster"))
			})

			It("should be rejected when snapshot timeout value is negative", func() {
				timeoutSeconds := int32(-1)
				clusterDef.Spec.Snapshot.TimeoutSeconds = &timeoutSeconds

				_, err := ACluster(clusterDef)
				Expect(err).To(MatchError("invalid snapshot timeoutSeconds value -1, must be non-negative for Cassandra cluster definition: mynamespace.mycluster"))
			})

			It("should be rejected when retention policy's retention period value is negative", func() {
				retentionPeriodDays := int32(-1)
				clusterDef.Spec.Snapshot.RetentionPolicy.RetentionPeriodDays = &retentionPeriodDays

				_, err := ACluster(clusterDef)
				Expect(err).To(MatchError("invalid snapshot retention policy retentionPeriodDays value -1, must be non-negative for Cassandra cluster definition: mynamespace.mycluster"))
			})

			It("should be rejected when retention policy's timeout value is negative", func() {
				cleanupTimeoutSeconds := int32(-1)
				clusterDef.Spec.Snapshot.RetentionPolicy.CleanupTimeoutSeconds = &cleanupTimeoutSeconds

				_, err := ACluster(clusterDef)
				Expect(err).To(MatchError("invalid snapshot retention policy cleanupTimeoutSeconds value -1, must be non-negative for Cassandra cluster definition: mynamespace.mycluster"))
			})

			It("should use the latest version of the cassandra snapshot image if one is not supplied for the cluster", func() {
				cluster, err := ACluster(clusterDef)
				Expect(err).ToNot(HaveOccurred())
				Expect(*cluster.definition.Spec.Snapshot.Image).To(Equal("skyuk/cassandra-snapshot:latest"))
			})

			It("should use the specified version of the cassandra snapshot image if one is given", func() {
				img := "somerepo/some-snapshot-image:v1.0"
				clusterDef.Spec.Snapshot.Image = &img
				cluster, err := ACluster(clusterDef)
				Expect(err).ToNot(HaveOccurred())
				Expect(*cluster.definition.Spec.Snapshot.Image).To(Equal("somerepo/some-snapshot-image:v1.0"))
			})

		})
	})
})

var _ = Describe("identification of custom config maps", func() {
	It("should look like a custom configmap when it ending with the correct suffix", func() {
		configMap := v1.ConfigMap{ObjectMeta: metaV1.ObjectMeta{Name: "cluster1-config"}}
		Expect(LooksLikeACassandraConfigMap(&configMap)).To(BeTrue())
	})

	It("should not look like a custom configmap when not ending with the correct suffix", func() {
		configMap := v1.ConfigMap{ObjectMeta: metaV1.ObjectMeta{Name: "cluster1-config-more"}}
		Expect(LooksLikeACassandraConfigMap(&configMap)).To(BeFalse())
	})

	It("should identify when a custom config map is not related to any cluster", func() {
		clusters := map[string]*Cluster{"cluster1": {definition: &v1alpha1.Cassandra{ObjectMeta: metaV1.ObjectMeta{Name: "cluster1"}}}}
		configMap := v1.ConfigMap{ObjectMeta: metaV1.ObjectMeta{Name: "cluster1-config-for-something-else"}}

		Expect(ConfigMapBelongsToAManagedCluster(clusters, &configMap)).To(BeFalse())
	})

	It("should derive the name of a cluster from a config map name", func() {
		configMap := v1.ConfigMap{ObjectMeta: metaV1.ObjectMeta{Name: "cluster1-config", Namespace: "the-namespace"}}
		clusterName, err := QualifiedClusterNameFor(&configMap)

		Expect(err).ToNot(HaveOccurred())
		Expect(clusterName).To(Equal("the-namespace.cluster1"))
	})

	It("should fail to derive the name of a cluster from a config map which does not fit the naming convention", func() {
		configMap := v1.ConfigMap{ObjectMeta: metaV1.ObjectMeta{Name: "cluster1-something-else", Namespace: "the-namespace"}}
		_, err := QualifiedClusterNameFor(&configMap)

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("creation of stateful sets", func() {
	var clusterDef *v1alpha1.Cassandra
	var configMap = &v1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "mycluster-config",
			Namespace: NAMESPACE,
		},
		Data: map[string]string{
			"test": "value",
		},
	}
	BeforeEach(func() {
		clusterDef = &v1alpha1.Cassandra{
			ObjectMeta: metaV1.ObjectMeta{Name: CLUSTER, Namespace: NAMESPACE},
			Spec: v1alpha1.CassandraSpec{
				Racks: []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}, {Name: "b", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}},
				Pod: v1alpha1.Pod{
					Memory:         resource.MustParse("1Gi"),
					CPU:            resource.MustParse("100m"),
					StorageSize:    resource.MustParse("1Gi"),
					ReadinessProbe: defaultReadinessProbe.DeepCopy(),
					LivenessProbe:  defaultLivenessProbe.DeepCopy(),
				},
			},
		}
	})

	It("should add init containers for config initialisation and bootstrapping", func() {
		// given
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		// when
		statefulSet := cluster.createStatefulSetForRack(&clusterDef.Spec.Racks[0], nil)

		// then
		Expect(statefulSet.Spec.Template.Spec.InitContainers).To(HaveLen(2))

		Expect(statefulSet.Spec.Template.Spec.InitContainers[0].Name).To(Equal("init-config"))
		Expect(statefulSet.Spec.Template.Spec.InitContainers[0].Image).To(Equal(*cluster.definition.Spec.Pod.Image))
		Expect(statefulSet.Spec.Template.Spec.InitContainers[0].Command).To(Equal([]string{"sh", "-c", "cp -vr /etc/cassandra/* /configuration"}))
		Expect(*statefulSet.Spec.Template.Spec.InitContainers[0].Resources.Requests.Memory()).To(Equal(clusterDef.Spec.Pod.Memory))
		Expect(*statefulSet.Spec.Template.Spec.InitContainers[0].Resources.Requests.Cpu()).To(Equal(clusterDef.Spec.Pod.CPU))
		Expect(*statefulSet.Spec.Template.Spec.InitContainers[0].Resources.Limits.Memory()).To(Equal(clusterDef.Spec.Pod.Memory))

		Expect(statefulSet.Spec.Template.Spec.InitContainers[1].Name).To(Equal("cassandra-bootstrapper"))
		Expect(statefulSet.Spec.Template.Spec.InitContainers[1].Image).To(ContainSubstring("skyuk/cassandra-bootstrapper:latest"))
		Expect(*statefulSet.Spec.Template.Spec.InitContainers[1].Resources.Requests.Memory()).To(Equal(clusterDef.Spec.Pod.Memory))
		Expect(*statefulSet.Spec.Template.Spec.InitContainers[1].Resources.Requests.Cpu()).To(Equal(clusterDef.Spec.Pod.CPU))
		Expect(*statefulSet.Spec.Template.Spec.InitContainers[1].Resources.Limits.Memory()).To(Equal(clusterDef.Spec.Pod.Memory))
	})

	It("should create the bootstrapper init container with the specified image if one is given", func() {
		clusterDef.Spec.Pod.BootstrapperImage = ptr.String("somerepo/abootstapperimage:v1")
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		statefulSet := cluster.createStatefulSetForRack(&clusterDef.Spec.Racks[0], nil)
		Expect(statefulSet.Spec.Template.Spec.InitContainers[1].Name).To(Equal("cassandra-bootstrapper"))
		Expect(statefulSet.Spec.Template.Spec.InitContainers[1].Image).To(Equal("somerepo/abootstapperimage:v1"))
	})

	It("should define environment variables for pod memory and cpu in bootstrapper init-container", func() {
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		statefulSet := cluster.createStatefulSetForRack(&clusterDef.Spec.Racks[0], nil)
		Expect(statefulSet.Spec.Template.Spec.InitContainers).To(HaveLen(2))
		Expect(statefulSet.Spec.Template.Spec.InitContainers[1].Env).To(ContainElement(v1.EnvVar{Name: "POD_CPU_MILLICORES", Value: "100"}))
		Expect(statefulSet.Spec.Template.Spec.InitContainers[1].Env).To(ContainElement(v1.EnvVar{Name: "POD_MEMORY_BYTES", Value: strconv.Itoa(1024 * 1024 * 1024)}))
	})

	It("should define environment variable for extra classpath in main container", func() {
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		statefulSet := cluster.createStatefulSetForRack(&clusterDef.Spec.Racks[0], nil)
		Expect(statefulSet.Spec.Template.Spec.Containers[0].Env).To(ContainElement(v1.EnvVar{Name: "EXTRA_CLASSPATH", Value: "/extra-lib/cassandra-seed-provider.jar"}))
	})

	It("should define emptyDir volumes for configuration and extra libraries", func() {
		// given
		cluster, err := ACluster(clusterDef)
		Expect(err).ToNot(HaveOccurred())

		// when
		statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

		// then
		volumes := statefulSet.Spec.Template.Spec.Volumes
		Expect(volumes).To(HaveLen(2))
		Expect(volumes).To(haveExactly(1, matchingEmptyDir("configuration")))
		Expect(volumes).To(haveExactly(1, matchingEmptyDir("extra-lib")))
	})

	It("should mount a persistent volume claim into the main container if useEmptyDir is not set", func() {
		// given
		cluster, err := ACluster(clusterDef)
		Expect(err).ToNot(HaveOccurred())

		// when
		statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

		// then
		Expect(statefulSet.Spec.Template.Spec.Volumes).To(haveExactly(0, matchingEmptyDir(fmt.Sprintf("cassandra-storage-%s", clusterDef.Name))))

		mainContainerVolumeMounts := statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts
		Expect(mainContainerVolumeMounts).To(HaveLen(3))
		Expect(mainContainerVolumeMounts).To(haveExactly(1, matchingVolumeMount(fmt.Sprintf("cassandra-storage-%s", clusterDef.Name), "/var/lib/cassandra")))
	})

	It("should mount an emptyDir into the main container if useEmptyDir is set", func() {
		// given
		clusterDef.Spec.UseEmptyDir = ptr.Bool(true)
		clusterDef.Spec.Pod.StorageSize = resource.MustParse("0")
		cluster, err := ACluster(clusterDef)
		Expect(err).ToNot(HaveOccurred())

		// when
		statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

		// then
		volumes := statefulSet.Spec.Template.Spec.Volumes
		Expect(volumes).To(HaveLen(3))
		Expect(volumes).To(haveExactly(1, matchingEmptyDir(fmt.Sprintf("cassandra-storage-%s", clusterDef.Name))))

		mainContainerVolumeMounts := statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts
		Expect(mainContainerVolumeMounts).To(HaveLen(3))
		Expect(mainContainerVolumeMounts).To(haveExactly(1, matchingVolumeMount(fmt.Sprintf("cassandra-storage-%s", clusterDef.Name), "/var/lib/cassandra")))
	})

	It("should mount the configuration and extra-lib volumes in the main container", func() {
		// given
		cluster, err := ACluster(clusterDef)
		Expect(err).ToNot(HaveOccurred())

		// when
		statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], configMap)

		// then
		mainContainerVolumeMounts := statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts
		Expect(mainContainerVolumeMounts).To(HaveLen(3))
		Expect(mainContainerVolumeMounts).To(haveExactly(1, matchingVolumeMount("configuration", "/etc/cassandra")))
		Expect(mainContainerVolumeMounts).To(haveExactly(1, matchingVolumeMount("extra-lib", "/extra-lib")))
	})

	Context("a cluster with a custom configMap is created", func() {
		It("should mount the configuration volume in the init-config container", func() {
			// given
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())

			// when
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], configMap)

			// then
			initConfigContainerVolumeMounts := statefulSet.Spec.Template.Spec.InitContainers[0].VolumeMounts
			Expect(initConfigContainerVolumeMounts).To(HaveLen(1))
			Expect(initConfigContainerVolumeMounts).To(haveExactly(1, matchingVolumeMount("configuration", "/configuration")))
		})

		It("should mount the configMap, configuration and extra-lib volumes in the bootstrap container", func() {
			// given
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())

			// when
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], configMap)

			// then
			volumes := statefulSet.Spec.Template.Spec.Volumes
			Expect(volumes).To(HaveLen(3))
			Expect(volumes).To(haveExactly(1, matchingConfigMap("cassandra-custom-config-mycluster", "mycluster-config")))

			bootstrapContainerVolumeMounts := statefulSet.Spec.Template.Spec.InitContainers[1].VolumeMounts
			Expect(bootstrapContainerVolumeMounts).To(HaveLen(3))
			Expect(bootstrapContainerVolumeMounts).To(haveExactly(1, matchingVolumeMount("configuration", "/configuration")))
			Expect(bootstrapContainerVolumeMounts).To(haveExactly(1, matchingVolumeMount("extra-lib", "/extra-lib")))
			Expect(bootstrapContainerVolumeMounts).To(haveExactly(1, matchingVolumeMount("cassandra-custom-config-mycluster", "/custom-config")))
		})
	})

	Context("a cluster without a custom configMap is created", func() {
		It("should not create the volume configMap and its corresponding mount in the bootstrap container", func() {
			// given
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())

			// when
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

			Expect(statefulSet.Spec.Template.Spec.Volumes).To(HaveLen(2))

			bootstrapContainerVolumeMounts := statefulSet.Spec.Template.Spec.InitContainers[1].VolumeMounts
			Expect(bootstrapContainerVolumeMounts).To(HaveLen(2))
			Expect(bootstrapContainerVolumeMounts).To(haveExactly(0, matchingVolumeMount("cassandra-custom-config-mycluster", "/custom-config")))
			Expect(bootstrapContainerVolumeMounts).To(haveExactly(1, matchingVolumeMount("configuration", "/configuration")))
			Expect(bootstrapContainerVolumeMounts).To(haveExactly(1, matchingVolumeMount("extra-lib", "/extra-lib")))
		})
	})
	Context("sidecar container", func() {
		It("configure the environment variables required by the cassandra-sidecar server", func() {
			// given
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

			// when
			actualEnv := statefulSet.Spec.Template.Spec.Containers[1].Env

			// then
			Expect(actualEnv).To(And(
				ContainElement(v1.EnvVar{Name: "NODE_LISTEN_ADDRESS", ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				}}),
				ContainElement(v1.EnvVar{Name: "CLUSTER_NAME", Value: cluster.Name()}),
				ContainElement(v1.EnvVar{Name: "CLUSTER_NAMESPACE", Value: cluster.Namespace()}),
			))
		})

		It("should use the maxSidecarCPURequest if that is lower than Cassandra.Spec.Pod.CPU", func() {
			// given
			clusterDef.Spec.Pod.CPU = resource.MustParse("2")
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

			// when
			actualResources := statefulSet.Spec.Template.Spec.Containers[1].Resources

			// then
			Expect(actualResources.Requests.Cpu()).To(Equal(&maxSidecarCPURequest))
		})

		It("should allow CPU bursting configurations", func() {
			// given
			clusterDef.Spec.Pod.CPU = resource.MustParse("0")
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

			// when
			actualResources := statefulSet.Spec.Template.Spec.Containers[1].Resources

			// then
			Expect(*actualResources.Requests.Cpu()).To(Equal(resource.MustParse("0")))
		})

		It("should use the maxSidecarMemoryRequest if that is lower than Cassandra.Spec.Pod.Memory", func() {
			// given
			clusterDef.Spec.Pod.Memory = resource.MustParse("1Ti")
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

			// when
			actualResources := statefulSet.Spec.Template.Spec.Containers[1].Resources

			// then
			Expect(actualResources.Requests.Memory()).To(Equal(&maxSidecarMemoryRequest))
		})

		It("should allow Memory bursting configurations", func() {
			// given
			clusterDef.Spec.Pod.Memory = resource.MustParse("1")
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

			// when
			actualResources := statefulSet.Spec.Template.Spec.Containers[1].Resources

			// then
			Expect(*actualResources.Requests.Memory()).To(Equal(resource.MustParse("1")))
		})
	})
})

var _ = Describe("modification of stateful sets", func() {
	var clusterDef *v1alpha1.Cassandra
	var configMap = &v1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "mycluster-config",
			Namespace: NAMESPACE,
		},
		Data: map[string]string{
			"test": "value",
		},
	}
	BeforeEach(func() {
		clusterDef = &v1alpha1.Cassandra{
			ObjectMeta: metaV1.ObjectMeta{Name: CLUSTER, Namespace: NAMESPACE},
			Spec: v1alpha1.CassandraSpec{
				Racks: []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}, {Name: "b", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}},
				Pod: v1alpha1.Pod{
					Memory:      resource.MustParse("1Gi"),
					CPU:         resource.MustParse("100m"),
					StorageSize: resource.MustParse("1Gi"),
				},
			},
		}
	})

	Context("the custom configMap is added", func() {
		It("should add the configMap volume and its corresponding mount to the cassandra-bootstrapper init-container", func() {
			// given
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

			// when
			err = cluster.AddCustomConfigVolumeToStatefulSet(statefulSet, configMap)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(statefulSet.Spec.Template.Spec.Volumes).To(HaveLen(3))
			Expect(statefulSet.Spec.Template.Spec.Volumes).To(haveExactly(1, matchingConfigMap("cassandra-custom-config-mycluster", "mycluster-config")))

			Expect(statefulSet.Spec.Template.Spec.InitContainers[1].VolumeMounts).To(HaveLen(3))
			Expect(statefulSet.Spec.Template.Spec.InitContainers[1].VolumeMounts).To(haveExactly(1, matchingVolumeMount("cassandra-custom-config-mycluster", "/custom-config")))
		})

		It("should add a config map hash annotation to the pod spec", func() {
			// given
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], nil)

			// when
			err = cluster.AddCustomConfigVolumeToStatefulSet(statefulSet, configMap)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(statefulSet.Spec.Template.Annotations[ConfigHashAnnotation]).ToNot(BeEmpty())
		})
	})

	Context("the custom configMap is removed", func() {
		It("should remove the configMap volume and its corresponding mount in the cassandra-bootstrapper init-container", func() {
			// given
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], configMap)

			// when
			err = cluster.RemoveCustomConfigVolumeFromStatefulSet(statefulSet, nil)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(statefulSet.Spec.Template.Spec.Volumes).To(HaveLen(2))
			Expect(statefulSet.Spec.Template.Spec.InitContainers[1].VolumeMounts).To(HaveLen(2))
			Expect(statefulSet.Spec.Template.Spec.InitContainers[1].VolumeMounts).To(haveExactly(1, matchingVolumeMount("configuration", "/configuration")))
			Expect(statefulSet.Spec.Template.Spec.InitContainers[1].VolumeMounts).To(haveExactly(1, matchingVolumeMount("extra-lib", "/extra-lib")))
		})

		It("should remove the config map hash annotation from the pod spec", func() {
			// given
			cluster, err := ACluster(clusterDef)
			Expect(err).ToNot(HaveOccurred())
			statefulSet := cluster.createStatefulSetForRack(&cluster.Racks()[0], configMap)

			// when
			err = cluster.RemoveCustomConfigVolumeFromStatefulSet(statefulSet, nil)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(statefulSet.Spec.Template.Annotations[ConfigHashAnnotation]).To(BeEmpty())
		})
	})
})

var _ = Describe("creation of snapshot job", func() {
	var (
		clusterDef      *v1alpha1.Cassandra
		snapshotTimeout = int32(10)
	)

	BeforeEach(func() {
		clusterDef = &v1alpha1.Cassandra{
			ObjectMeta: metaV1.ObjectMeta{Name: CLUSTER, Namespace: NAMESPACE},
			Spec: v1alpha1.CassandraSpec{
				Racks: []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}, {Name: "b", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}},
				Pod: v1alpha1.Pod{
					Memory:      resource.MustParse("1Gi"),
					CPU:         resource.MustParse("100m"),
					StorageSize: resource.MustParse("1Gi"),
				},
				Snapshot: &v1alpha1.Snapshot{
					Schedule:       "01 23 * * *",
					TimeoutSeconds: &snapshotTimeout,
				},
			},
		}
	})

	It("should create a cronjob named after the cluster that will trigger at the specified schedule", func() {
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotJob()
		Expect(cronJob.Name).To(Equal(fmt.Sprintf("%s-snapshot", clusterDef.Name)))
		Expect(cronJob.Namespace).To(Equal(clusterDef.Namespace))
		Expect(cronJob.Labels).To(And(
			HaveKeyWithValue(OperatorLabel, clusterDef.Name),
			HaveKeyWithValue("app", fmt.Sprintf("%s-snapshot", clusterDef.Name)),
		))
		Expect(cronJob.Spec.Schedule).To(Equal("01 23 * * *"))
		Expect(cronJob.Spec.ConcurrencyPolicy).To(Equal(v1beta1.ForbidConcurrent))
	})

	It("should create a cronjob with its associated job named after the cluster in the same namespace", func() {
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotJob()
		backupJob := cronJob.Spec.JobTemplate
		Expect(backupJob.Name).To(Equal(fmt.Sprintf("%s-snapshot", clusterDef.Name)))
		Expect(backupJob.Namespace).To(Equal(clusterDef.Namespace))
		Expect(backupJob.Labels).To(And(
			HaveKeyWithValue(OperatorLabel, clusterDef.Name),
			HaveKeyWithValue("app", fmt.Sprintf("%s-snapshot", clusterDef.Name)),
		))

		backupPod := cronJob.Spec.JobTemplate.Spec.Template
		Expect(backupPod.Name).To(Equal(fmt.Sprintf("%s-snapshot", clusterDef.Name)))
		Expect(backupPod.Namespace).To(Equal(clusterDef.Namespace))
		Expect(backupPod.Labels).To(And(
			HaveKeyWithValue(OperatorLabel, clusterDef.Name),
			HaveKeyWithValue("app", fmt.Sprintf("%s-snapshot", clusterDef.Name)),
		))
	})

	It("should create a cronjob that will trigger a snapshot creation for the whole cluster when no keyspace specified", func() {
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotJob()
		Expect(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers).To(HaveLen(1))

		snapshotContainer := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
		Expect(snapshotContainer.Name).To(Equal(fmt.Sprintf("%s-snapshot", clusterDef.Name)))
		Expect(snapshotContainer.Command).To(Equal([]string{
			"/cassandra-snapshot", "create",
			"-n", cluster.Namespace(),
			"-l", fmt.Sprintf("%s=%s,%s=%s", OperatorLabel, clusterDef.Name, "app", clusterDef.Name),
			"-t", durationSeconds(&snapshotTimeout).String(),
		}))
		Expect(snapshotContainer.Image).To(ContainSubstring("skyuk/cassandra-snapshot:latest"))
	})

	It("should create a cronjob that will trigger a snapshot creation for the specified keyspaces", func() {
		clusterDef.Spec.Snapshot.Keyspaces = []string{"keyspace1", "keyspace50"}
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotJob()
		Expect(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers).To(HaveLen(1))

		snapshotContainer := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
		Expect(snapshotContainer.Name).To(Equal(fmt.Sprintf("%s-snapshot", clusterDef.Name)))
		Expect(snapshotContainer.Command).To(Equal([]string{
			"/cassandra-snapshot", "create",
			"-n", cluster.Namespace(),
			"-l", fmt.Sprintf("%s=%s,%s=%s", OperatorLabel, clusterDef.Name, "app", clusterDef.Name),
			"-t", durationSeconds(&snapshotTimeout).String(),
			"-k", "keyspace1,keyspace50",
		}))
		Expect(snapshotContainer.Image).To(ContainSubstring("skyuk/cassandra-snapshot:latest"))
	})

	It("should create a cronjob which pod will restart in case of failure", func() {
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotJob()

		snapshotPod := cronJob.Spec.JobTemplate.Spec.Template.Spec
		Expect(snapshotPod.RestartPolicy).To(Equal(v1.RestartPolicyOnFailure))
	})

	It("should not pass a snapshot time to the snapshot command if none is specified in the cluster spec", func() {
		clusterDef.Spec.Snapshot.TimeoutSeconds = nil
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotJob()
		Expect(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers).To(HaveLen(1))

		snapshotContainer := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
		Expect(snapshotContainer.Command).To(Equal([]string{
			"/cassandra-snapshot", "create",
			"-n", cluster.Namespace(),
			"-l", fmt.Sprintf("%s=%s,%s=%s", OperatorLabel, clusterDef.Name, "app", clusterDef.Name),
		}))
	})

	It("should not create a snapshot job if none is specified in the cluster spec", func() {
		clusterDef.Spec.Snapshot = nil
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotJob()

		Expect(cronJob).To(BeNil())
	})

	It("should create a cronjob which pod is using the specified snapshot image", func() {
		img := "somerepo/snapshot:v1"
		clusterDef.Spec.Snapshot.Image = &img
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotJob()

		snapshotContainer := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
		Expect(snapshotContainer.Image).To(ContainSubstring("somerepo/snapshot:v1"))
	})

})

var _ = Describe("creation of snapshot cleanup job", func() {
	var (
		clusterDef      *v1alpha1.Cassandra
		snapshotTimeout = int32(10)
		cleanupTimeout  = int32(5)
		retentionPeriod = int32(1)
	)

	BeforeEach(func() {
		clusterDef = &v1alpha1.Cassandra{
			ObjectMeta: metaV1.ObjectMeta{Name: CLUSTER, Namespace: NAMESPACE},
			Spec: v1alpha1.CassandraSpec{
				Racks: []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}, {Name: "b", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}},
				Pod: v1alpha1.Pod{
					Memory:      resource.MustParse("1Gi"),
					CPU:         resource.MustParse("100m"),
					StorageSize: resource.MustParse("1Gi"),
				},
				Snapshot: &v1alpha1.Snapshot{
					Schedule:       "1 23 * * *",
					TimeoutSeconds: &snapshotTimeout,
					RetentionPolicy: &v1alpha1.RetentionPolicy{
						Enabled:               ptr.Bool(true),
						RetentionPeriodDays:   &retentionPeriod,
						CleanupSchedule:       "0 9 * * *",
						CleanupTimeoutSeconds: &cleanupTimeout,
					},
				},
			},
		}
	})

	It("should not create a cleanup job if no retention policy is specified in the cluster spec", func() {
		clusterDef.Spec.Snapshot.RetentionPolicy = nil
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotCleanupJob()

		Expect(cronJob).To(BeNil())
	})

	It("should not create a cleanup job if the retention policy is disabled in the cluster spec", func() {
		clusterDef.Spec.Snapshot.RetentionPolicy = &v1alpha1.RetentionPolicy{
			Enabled: ptr.Bool(false),
		}
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotCleanupJob()

		Expect(cronJob).To(BeNil())
	})

	It("should create a cronjob named after the cluster that will trigger at the specified schedule", func() {
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotCleanupJob()
		Expect(cronJob.Name).To(Equal(fmt.Sprintf("%s-snapshot-cleanup", clusterDef.Name)))
		Expect(cronJob.Namespace).To(Equal(clusterDef.Namespace))
		Expect(cronJob.Labels).To(And(
			HaveKeyWithValue(OperatorLabel, clusterDef.Name),
			HaveKeyWithValue("app", fmt.Sprintf("%s-snapshot-cleanup", clusterDef.Name)),
		))
		Expect(cronJob.Spec.Schedule).To(Equal("0 9 * * *"))
	})

	It("should create a cronjob with its associated job named after the cluster in the same namespace", func() {
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotCleanupJob()
		cleanupJob := cronJob.Spec.JobTemplate
		Expect(cleanupJob.Name).To(Equal(fmt.Sprintf("%s-snapshot-cleanup", clusterDef.Name)))
		Expect(cleanupJob.Namespace).To(Equal(clusterDef.Namespace))
		Expect(cleanupJob.Labels).To(And(
			HaveKeyWithValue(OperatorLabel, clusterDef.Name),
			HaveKeyWithValue("app", fmt.Sprintf("%s-snapshot-cleanup", clusterDef.Name)),
		))

		cleanupPod := cronJob.Spec.JobTemplate.Spec.Template
		Expect(cleanupPod.Name).To(Equal(fmt.Sprintf("%s-snapshot-cleanup", clusterDef.Name)))
		Expect(cleanupPod.Namespace).To(Equal(clusterDef.Namespace))
		Expect(cleanupPod.Labels).To(And(
			HaveKeyWithValue(OperatorLabel, clusterDef.Name),
			HaveKeyWithValue("app", fmt.Sprintf("%s-snapshot-cleanup", clusterDef.Name)),
		))
	})

	It("should create a cronjob that will trigger a snapshot cleanup", func() {
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotCleanupJob()
		Expect(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers).To(HaveLen(1))

		cleanupContainer := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
		Expect(cleanupContainer.Name).To(Equal(fmt.Sprintf("%s-snapshot-cleanup", clusterDef.Name)))
		Expect(cleanupContainer.Command).To(Equal([]string{
			"/cassandra-snapshot", "cleanup",
			"-n", cluster.Namespace(),
			"-l", fmt.Sprintf("%s=%s,%s=%s", OperatorLabel, clusterDef.Name, "app", clusterDef.Name),
			"-r", durationDays(&retentionPeriod).String(),
			"-t", durationSeconds(&cleanupTimeout).String(),
		}))
		Expect(cleanupContainer.Image).To(ContainSubstring("skyuk/cassandra-snapshot:latest"))
	})

	It("should create a cronjob that will trigger a snapshot cleanup without explicit retention period", func() {
		clusterDef.Spec.Snapshot.RetentionPolicy.RetentionPeriodDays = nil
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotCleanupJob()
		Expect(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers).To(HaveLen(1))

		cleanupContainer := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
		Expect(cleanupContainer.Name).To(Equal(fmt.Sprintf("%s-snapshot-cleanup", clusterDef.Name)))
		Expect(cleanupContainer.Command).To(Equal([]string{
			"/cassandra-snapshot", "cleanup",
			"-n", cluster.Namespace(),
			"-l", fmt.Sprintf("%s=%s,%s=%s", OperatorLabel, clusterDef.Name, "app", clusterDef.Name),
			"-t", durationSeconds(&cleanupTimeout).String(),
		}))
		Expect(cleanupContainer.Image).To(ContainSubstring("skyuk/cassandra-snapshot:latest"))
	})

	It("should create a cronjob which pod is using the specified snapshot image", func() {
		img := "somerepo/snapshot:v1"
		clusterDef.Spec.Snapshot.Image = &img
		cluster, err := ACluster(clusterDef)
		Expect(err).NotTo(HaveOccurred())

		cronJob := cluster.CreateSnapshotCleanupJob()

		cleanupContainer := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
		Expect(cleanupContainer.Image).To(ContainSubstring("somerepo/snapshot:v1"))
	})

})

func ACluster(clusterDef *v1alpha1.Cassandra) (*Cluster, error) {
	return New(clusterDef)
}

//

func haveExactly(count int, subMatcher types.GomegaMatcher) types.GomegaMatcher {
	return &haveExactlyMatcher{count, subMatcher}
}

type haveExactlyMatcher struct {
	count      int
	subMatcher types.GomegaMatcher
}

func (h *haveExactlyMatcher) Match(actual interface{}) (success bool, err error) {
	arr := reflect.ValueOf(actual)

	if arr.Kind() != reflect.Slice {
		return false, fmt.Errorf("expected []interface{}, got %v", arr.Kind())
	}

	if arr.Len() == 0 {
		fmt.Printf("zero-length slice")
		return false, fmt.Errorf("zero-length slice")
	}

	matching := 0
	for i := 0; i < arr.Len(); i++ {
		item := arr.Index(i).Interface()
		if success, _ := h.subMatcher.Match(item); success {
			matching++
		}
	}

	return matching == h.count, nil
}

func (h *haveExactlyMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("expected exactly one element of %v to match %v", actual, h.subMatcher)
}

func (h *haveExactlyMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("did not expect exactly one element of %v to match %v", actual, h.subMatcher)
}

//

func matchingConfigMap(volumeName, localObjectReference string) types.GomegaMatcher {
	return &configMapMatcher{volumeName, localObjectReference}
}

type configMapMatcher struct {
	volumeName           string
	localObjectReference string
}

func (h *configMapMatcher) Match(actual interface{}) (success bool, err error) {
	switch v := actual.(type) {
	case v1.Volume:
		return v.Name == h.volumeName && v.ConfigMap != nil && v.ConfigMap.LocalObjectReference.Name == h.localObjectReference, nil
	default:
		return false, fmt.Errorf("expected v1.Volume, got %v", actual)
	}
}

func (h *configMapMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("expected volume with name %s referencing config map %s", h.volumeName, h.localObjectReference)
}

func (h *configMapMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("did not expect volume with name %s referencing config map %s", h.volumeName, h.localObjectReference)
}

//

func matchingEmptyDir(volumeName string) types.GomegaMatcher {
	return &emptyDirMatcher{volumeName}
}

type emptyDirMatcher struct {
	volumeName string
}

func (h *emptyDirMatcher) Match(actual interface{}) (success bool, err error) {
	switch v := actual.(type) {
	case v1.Volume:
		return v.Name == h.volumeName && v.EmptyDir != nil, nil
	default:
		return false, fmt.Errorf("expected v1.Volume, got %v", actual)
	}
}

func (h *emptyDirMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("expected emptyDir volume with name %s", h.volumeName)
}

func (h *emptyDirMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("did not expect emptyDir volume with name %s", h.volumeName)
}

//

func matchingVolumeMount(mount, path string) types.GomegaMatcher {
	return &volumeMountMatcher{mount, path}
}

type volumeMountMatcher struct {
	mount string
	path  string
}

func (h *volumeMountMatcher) Match(actual interface{}) (success bool, err error) {
	switch m := actual.(type) {
	case v1.VolumeMount:
		return m.Name == h.mount && m.MountPath == h.path, nil
	default:
		return false, fmt.Errorf("expected v1.VolumeMount, got %v", actual)
	}
}

func (h *volumeMountMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("expected volume mount with name %s and path %s", h.mount, h.path)
}

func (h *volumeMountMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("did not expect volume mount with name %s and path %s", h.mount, h.path)
}

var _ = Describe("utility functions", func() {
	DescribeTable(
		"minQuantity",
		func(q1, q2, q3 string) {
			actual := minQuantity(resource.MustParse(q1), resource.MustParse(q2))
			expected := resource.MustParse(q3)
			Expect(actual).To(Equal(expected))
		},
		Entry("q1 > q2", "2Mi", "1Mi", "1Mi"),
		Entry("q1 < q2", "300m", "200m", "200m"),
		Entry("q1 == q2", "4", "4000m", "4"),
		Entry("q1 == q2 (retain scale)", "4000m", "4", "4000m"),
	)
})
