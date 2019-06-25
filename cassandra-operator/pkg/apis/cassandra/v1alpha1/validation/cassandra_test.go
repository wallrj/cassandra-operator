package validation_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1/validation"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Cluster Suite", test.CreateParallelReporters("cluster"))
}

var _ = Describe("validation", func() {
	Context("ValidateCassandra", func() {
		var (
			cass *v1alpha1.Cassandra
		)

		BeforeEach(func() {
			cass = &v1alpha1.Cassandra{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cluster1",
					Namespace: "ns1",
				},
				Spec: v1alpha1.CassandraSpec{
					UseEmptyDir: ptr.Bool(false),
					Racks: []v1alpha1.Rack{
						{
							Name:         "rack1",
							Zone:         "zone1",
							StorageClass: "fast",
							Replicas:     1,
						},
					},
					Pod: v1alpha1.Pod{
						CPU:         resource.MustParse("0"),
						Memory:      resource.MustParse("1Gi"),
						StorageSize: resource.MustParse("100Gi"),
						LivenessProbe: &v1alpha1.Probe{
							FailureThreshold:    ptr.Int32(1),
							InitialDelaySeconds: ptr.Int32(1),
							PeriodSeconds:       ptr.Int32(1),
							SuccessThreshold:    ptr.Int32(1),
							TimeoutSeconds:      ptr.Int32(1),
						},
						ReadinessProbe: &v1alpha1.Probe{
							FailureThreshold:    ptr.Int32(1),
							InitialDelaySeconds: ptr.Int32(1),
							PeriodSeconds:       ptr.Int32(1),
							SuccessThreshold:    ptr.Int32(1),
							TimeoutSeconds:      ptr.Int32(1),
						},
					},
					Snapshot: &v1alpha1.Snapshot{
						Schedule:       "1 23 * * *",
						TimeoutSeconds: ptr.Int32(1),
						RetentionPolicy: &v1alpha1.RetentionPolicy{
							RetentionPeriodDays:   ptr.Int32(1),
							CleanupTimeoutSeconds: ptr.Int32(0),
							CleanupSchedule:       "1 23 * * *",
						},
					},
				},
			}
		})

		Context("success cases", func() {
			AfterEach(func() {
				err := validation.ValidateCassandra(cass).ToAggregate()
				Expect(err).ToNot(HaveOccurred())
			})
			It("succeeds with a fully populated Cassandra object", func() {})
			It("succeeds without a Snapshot", func() {
				cass.Spec.Snapshot = nil
			})
			It("succeeds without a Snapshot.TimeoutSeconds", func() {
				cass.Spec.Snapshot.TimeoutSeconds = nil
			})
			It("succeeds without a Snapshot.RetentionPolicy", func() {
				cass.Spec.Snapshot.RetentionPolicy = nil
			})
			It("succeeds without a Snapshot.RetentionPolicy.RetentionPeriodDays", func() {
				cass.Spec.Snapshot.RetentionPolicy.RetentionPeriodDays = nil
			})
			It("succeeds without a Snapshot.RetentionPolicy.CleanupTimeoutSeconds", func() {
				cass.Spec.Snapshot.RetentionPolicy.CleanupTimeoutSeconds = nil
			})
		})

		Context("failure cases", func() {
			AfterEach(func() {
				err := validation.ValidateCassandra(cass).ToAggregate()
				fmt.Fprintf(GinkgoWriter, "INFO: Error message was: %q", err)
				Expect(err).To(HaveOccurred())
				t := CurrentGinkgoTestDescription()
				Expect(err.Error()).To(Equal(t.TestText))
			})

			Context("Spec", func() {
				Context("Snapshot", func() {
					It("fails if Schedule is empty", func() {
						cass.Spec.Snapshot.Schedule = ""
					})
					It("fails if Schedule is not valid cron syntax", func() {
						cass.Spec.Snapshot.Schedule = "x y z"
					})
					It("fails if TimeoutSeconds is < 0", func() {
						cass.Spec.Snapshot.TimeoutSeconds = ptr.Int32(-1)
					})
					Context("RetentionPolicy", func() {
						It("fails if RetentionPeriodDays is < 0", func() {
							cass.Spec.Snapshot.RetentionPolicy.RetentionPeriodDays = ptr.Int32(-1)
						})
						It("fails if CleanupTimeoutSeconds is < 0", func() {
							cass.Spec.Snapshot.RetentionPolicy.CleanupTimeoutSeconds = ptr.Int32(-1)
						})
						It("fails if CleanupSchedule is empty", func() {
							cass.Spec.Snapshot.RetentionPolicy.CleanupSchedule = ""
						})
						It("fails if CleanupSchedule is not valid cron syntax", func() {
							cass.Spec.Snapshot.RetentionPolicy.CleanupSchedule = "x y z"
						})
					})
				})
			})
		})
	})
})

func cass() *v1alpha1.Cassandra {
	return &v1alpha1.Cassandra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster1",
			Namespace: "ns1",
		},
		Spec: v1alpha1.CassandraSpec{
			UseEmptyDir: ptr.Bool(false),
			Racks: []v1alpha1.Rack{
				{
					Name:         "rack1",
					Zone:         "zone1",
					StorageClass: "fast",
					Replicas:     1,
				},
			},
			Pod: v1alpha1.Pod{
				CPU:         resource.MustParse("0"),
				Memory:      resource.MustParse("1Gi"),
				StorageSize: resource.MustParse("100Gi"),
				LivenessProbe: &v1alpha1.Probe{
					FailureThreshold:    ptr.Int32(1),
					InitialDelaySeconds: ptr.Int32(1),
					PeriodSeconds:       ptr.Int32(1),
					SuccessThreshold:    ptr.Int32(1),
					TimeoutSeconds:      ptr.Int32(1),
				},
				ReadinessProbe: &v1alpha1.Probe{
					FailureThreshold:    ptr.Int32(1),
					InitialDelaySeconds: ptr.Int32(1),
					PeriodSeconds:       ptr.Int32(1),
					SuccessThreshold:    ptr.Int32(1),
					TimeoutSeconds:      ptr.Int32(1),
				},
			},
			Snapshot: &v1alpha1.Snapshot{
				Schedule:       "1 23 * * *",
				TimeoutSeconds: ptr.Int32(1),
				RetentionPolicy: &v1alpha1.RetentionPolicy{
					RetentionPeriodDays:   ptr.Int32(1),
					CleanupTimeoutSeconds: ptr.Int32(0),
					CleanupSchedule:       "1 23 * * *",
				},
			},
		},
	}
}

var _ = Describe("validation2", func() {
	DescribeTable(
		"failure cases",
		func(expectedMessage string, mutate func(*v1alpha1.Cassandra)) {
			c := cass()
			mutate(c)
			err := validation.ValidateCassandra(c).ToAggregate()
			Expect(err).To(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "INFO: Error message was: %q\n", err)
			Expect(err.Error()).To(Equal(expectedMessage))
		},
		Entry(
			"Name is required",
			"metadata.name: Required value: name or generateName is required",
			func(c *v1alpha1.Cassandra) {
				c.Name = ""
			},
		),
		Entry(
			"Namespace is required",
			"metadata.namespace: Required value",
			func(c *v1alpha1.Cassandra) {
				c.Namespace = ""
			},
		),
		Entry(
			"Spec.Racks is required",
			"spec.Racks: Required value",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Racks = nil
			},
		),
		Entry(
			"Rack.Replicas must be positive",
			"spec.Racks.rack1.Replicas: Invalid value: -1: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Racks[0].Replicas = -1
			},
		),
		Entry(
			"Rack.Replicas must be >= 1",
			"spec.Racks.rack1.Replicas: Invalid value: 0: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Racks[0].Replicas = 0
			},
		),
		Entry(
			"Rack.StorageClass must not be empty when UseEmptyDir=false",
			"spec.Racks.rack1.StorageClass: Required value: because spec.useEmptyDir is true",
			func(c *v1alpha1.Cassandra) {
				c.Spec.UseEmptyDir = ptr.Bool(false)
				c.Spec.Racks[0].StorageClass = ""
			},
		),
		Entry(
			"Rack.StorageClass must not be empty when UseEmptyDir=false",
			"spec.Racks.rack1.StorageClass: Required value: because spec.useEmptyDir is true",
			func(c *v1alpha1.Cassandra) {
				c.Spec.UseEmptyDir = ptr.Bool(false)
				c.Spec.Racks[0].StorageClass = ""
			},
		),
		Entry(
			"Rack.Zone must not be empty when UseEmptyDir=false",
			"spec.Racks.rack1.Zone: Required value: because spec.useEmptyDir is true",
			func(c *v1alpha1.Cassandra) {
				c.Spec.UseEmptyDir = ptr.Bool(false)
				c.Spec.Racks[0].Zone = ""
			},
		),
		Entry(
			"Pod.Memory must be > 0",
			"spec.Pod.Memory: Invalid value: \"0\": must be > 0",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.Memory.Set(0)
			},
		),
		Entry(
			"Pod.StorageSize must be > 0 when UseEmptyDir=false",
			"spec.Pod.StorageSize: Invalid value: \"0\": must be > 0 when spec.useEmptyDir is false",
			func(c *v1alpha1.Cassandra) {
				c.Spec.UseEmptyDir = ptr.Bool(false)
				c.Spec.Pod.StorageSize.Set(0)
			},
		),
		Entry(
			"Pod.StorageSize must be 0 when UseEmptyDir=true",
			"spec.Pod.StorageSize: Invalid value: \"1\": must be 0 when spec.useEmptyDir is true",
			func(c *v1alpha1.Cassandra) {
				c.Spec.UseEmptyDir = ptr.Bool(true)
				c.Spec.Pod.StorageSize.Set(1)
			},
		),

		Entry(
			"LivenessProbe.FailureThreshold must be positive",
			"spec.Pod.LivenessProbe.FailureThreshold: Invalid value: -1: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.LivenessProbe.FailureThreshold = ptr.Int32(-1)
			},
		),
		Entry(
			"LivenessProbe.FailureThreshold must be >= 1",
			"spec.Pod.LivenessProbe.FailureThreshold: Invalid value: 0: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.LivenessProbe.FailureThreshold = ptr.Int32(0)
			},
		),
		Entry(
			"LivenessProbe.InitialDelaySeconds must be positive",
			"spec.Pod.LivenessProbe.InitialDelaySeconds: Invalid value: -1: must be >= 0",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.LivenessProbe.InitialDelaySeconds = ptr.Int32(-1)
			},
		),
		Entry(
			"LivenessProbe.PeriodSeconds must be positive",
			"spec.Pod.LivenessProbe.PeriodSeconds: Invalid value: -1: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.LivenessProbe.PeriodSeconds = ptr.Int32(-1)
			},
		),
		Entry(
			"LivenessProbe.PeriodSeconds must be >= 1",
			"spec.Pod.LivenessProbe.PeriodSeconds: Invalid value: 0: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.LivenessProbe.PeriodSeconds = ptr.Int32(0)
			},
		),
		Entry(
			"LivenessProbe.SuccessThreshold must be positive",
			"spec.Pod.LivenessProbe.SuccessThreshold: Invalid value: -1: must be 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.LivenessProbe.SuccessThreshold = ptr.Int32(-1)
			},
		),
		Entry(
			"LivenessProbe.SuccessThreshold must be 1 (<1)",
			"spec.Pod.LivenessProbe.SuccessThreshold: Invalid value: 0: must be 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.LivenessProbe.SuccessThreshold = ptr.Int32(0)
			},
		),
		Entry(
			"LivenessProbe.SuccessThreshold must be 1 (>1)",
			"spec.Pod.LivenessProbe.SuccessThreshold: Invalid value: 2: must be 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.LivenessProbe.SuccessThreshold = ptr.Int32(2)
			},
		),
		Entry(
			"LivenessProbe.TimeoutSeconds must be positive",
			"spec.Pod.LivenessProbe.TimeoutSeconds: Invalid value: -1: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.LivenessProbe.TimeoutSeconds = ptr.Int32(-1)
			},
		),
		Entry(
			"LivenessProbe.TimeoutSeconds must be >= 1",
			"spec.Pod.LivenessProbe.TimeoutSeconds: Invalid value: 0: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.LivenessProbe.TimeoutSeconds = ptr.Int32(0)
			},
		),

		Entry(
			"ReadinessProbe.FailureThreshold must be positive",
			"spec.Pod.ReadinessProbe.FailureThreshold: Invalid value: -1: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.ReadinessProbe.FailureThreshold = ptr.Int32(-1)
			},
		),
		Entry(
			"ReadinessProbe.FailureThreshold must be >= 1",
			"spec.Pod.ReadinessProbe.FailureThreshold: Invalid value: 0: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.ReadinessProbe.FailureThreshold = ptr.Int32(0)
			},
		),
		Entry(
			"ReadinessProbe.InitialDelaySeconds must be positive",
			"spec.Pod.ReadinessProbe.InitialDelaySeconds: Invalid value: -1: must be >= 0",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.ReadinessProbe.InitialDelaySeconds = ptr.Int32(-1)
			},
		),
		Entry(
			"ReadinessProbe.PeriodSeconds must be positive",
			"spec.Pod.ReadinessProbe.PeriodSeconds: Invalid value: -1: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.ReadinessProbe.PeriodSeconds = ptr.Int32(-1)
			},
		),
		Entry(
			"ReadinessProbe.PeriodSeconds must be >= 1",
			"spec.Pod.ReadinessProbe.PeriodSeconds: Invalid value: 0: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.ReadinessProbe.PeriodSeconds = ptr.Int32(0)
			},
		),
		Entry(
			"ReadinessProbe.SuccessThreshold must be positive",
			"spec.Pod.ReadinessProbe.SuccessThreshold: Invalid value: -1: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.ReadinessProbe.SuccessThreshold = ptr.Int32(-1)
			},
		),
		Entry(
			"ReadinessProbe.SuccessThreshold must be >= 1",
			"spec.Pod.ReadinessProbe.SuccessThreshold: Invalid value: 0: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.ReadinessProbe.SuccessThreshold = ptr.Int32(0)
			},
		),
		// Entry(
		//	"ReadinessProbe.SuccessThreshold must be 1 (>1)",
		//	"spec.Pod.ReadinessProbe.SuccessThreshold: Invalid value: 2: must be 1",
		//	func(c *v1alpha1.Cassandra) {
		//		c.Spec.Pod.ReadinessProbe.SuccessThreshold = ptr.Int32(2)
		//	},
		// ),
		Entry(
			"ReadinessProbe.TimeoutSeconds must be positive",
			"spec.Pod.ReadinessProbe.TimeoutSeconds: Invalid value: -1: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.ReadinessProbe.TimeoutSeconds = ptr.Int32(-1)
			},
		),
		Entry(
			"ReadinessProbe.TimeoutSeconds must be >= 1",
			"spec.Pod.ReadinessProbe.TimeoutSeconds: Invalid value: 0: must be >= 1",
			func(c *v1alpha1.Cassandra) {
				c.Spec.Pod.ReadinessProbe.TimeoutSeconds = ptr.Int32(0)
			},
		),
	)
})
