package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
)

func TestTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Types Suite", test.CreateParallelReporters("types"))
}

var _ = Describe("Cassandra Types", func() {
	Context("Pod spec equality", func() {
		var pod, otherPod *Pod

		BeforeEach(func() {
			pod = &Pod{
				BootstrapperImage: ptr.String("BootstrapperImage"),
				SidecarImage: ptr.String("SidecarImage"),
				Image: ptr.String("Image"),
				Memory: resource.MustParse("2Gi"),
				CPU: resource.MustParse("1000m"),
				StorageSize: resource.MustParse("1Gi"),
				LivenessProbe: &Probe{
					SuccessThreshold:    ptr.Int32(1),
					PeriodSeconds:       ptr.Int32(2),
					InitialDelaySeconds: ptr.Int32(3),
					FailureThreshold:    ptr.Int32(4),
					TimeoutSeconds:      ptr.Int32(5),
				},
				ReadinessProbe: &Probe{
					SuccessThreshold:    ptr.Int32(1),
					PeriodSeconds:       ptr.Int32(2),
					InitialDelaySeconds: ptr.Int32(3),
					FailureThreshold:    ptr.Int32(4),
					TimeoutSeconds:      ptr.Int32(5),
				},
			}
			otherPod = pod.DeepCopy()
		})

		Context("equality", func() {
			It("should be equal if all fields are equal", func() {
				Expect(podsEqual(pod, otherPod)).To(BeTrue())
			})

			It("should be equal when cpu value is the same but using a different amount", func() {
				pod.CPU = resource.MustParse("1")
				Expect(podsEqual(pod, otherPod)).To(BeTrue())
			})

			It("should be equal when memory value is the same but using a different amount", func() {
				pod.Memory = resource.MustParse("2048Mi")
				Expect(podsEqual(pod, otherPod)).To(BeTrue())
			})

			It("should be equal when storage size value is the same but using a different amount", func() {
				pod.StorageSize = resource.MustParse("1024Mi")
				Expect(podsEqual(pod, otherPod)).To(BeTrue())
			})
		})

		Context("inequality", func() {

			It("should not be equal when one pod is an empty struct", func() {
				Expect(pod.Equal(Pod{})).To(BeFalse())
				Expect(Pod{}.Equal(*pod)).To(BeFalse())
			})

			It("should not be equal when one pod has a nil bootstrap image", func() {
				pod.BootstrapperImage = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when pods have different bootstrap images", func() {
				pod.BootstrapperImage = ptr.String("another image")
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when one pod has a nil sidecar image", func() {
				pod.SidecarImage = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when pods have different sidecar images", func() {
				pod.SidecarImage = ptr.String("another image")
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when one pod has a nil cassandra image", func() {
				pod.Image = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when pods have different cassandra images", func() {
				pod.Image = ptr.String("another image")
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when one pod has no storage size", func() {
				pod.StorageSize = resource.Quantity{}
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when pods have different storage sizes", func() {
				pod.StorageSize = resource.MustParse("10Gi")
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when one pod has no cpu", func() {
				pod.CPU = resource.Quantity{}
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when pods have different number of cpu", func() {
				pod.CPU = resource.MustParse("10")
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when one pod has no memory", func() {
				pod.Memory = resource.Quantity{}
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when pods have different memory sizes", func() {
				pod.Memory = resource.MustParse("10")
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when liveness probes have different values", func() {
				pod.LivenessProbe.SuccessThreshold = ptr.Int32(20)
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.LivenessProbe.TimeoutSeconds = ptr.Int32(20)
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.LivenessProbe.FailureThreshold = ptr.Int32(20)
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.LivenessProbe.InitialDelaySeconds = ptr.Int32(20)
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.LivenessProbe.PeriodSeconds = ptr.Int32(20)
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when one liveness probe  has nil values", func() {
				pod.LivenessProbe.SuccessThreshold = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.LivenessProbe.TimeoutSeconds = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.LivenessProbe.FailureThreshold = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.LivenessProbe.InitialDelaySeconds = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.LivenessProbe.PeriodSeconds = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when readiness probes have different values", func() {
				pod.ReadinessProbe.SuccessThreshold = ptr.Int32(20)
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.ReadinessProbe.TimeoutSeconds = ptr.Int32(20)
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.ReadinessProbe.FailureThreshold = ptr.Int32(20)
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.ReadinessProbe.InitialDelaySeconds = ptr.Int32(20)
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.ReadinessProbe.PeriodSeconds = ptr.Int32(20)
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

			It("should not be equal when one readiness probe  has nil values", func() {
				pod.ReadinessProbe.SuccessThreshold = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.ReadinessProbe.TimeoutSeconds = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.ReadinessProbe.FailureThreshold = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.ReadinessProbe.InitialDelaySeconds = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())

				pod.ReadinessProbe.PeriodSeconds = nil
				Expect(podsNotEqual(pod, otherPod)).To(BeTrue())
			})

		})
	})

})


func podsEqual(pod, otherPod *Pod) bool {
	return pod.Equal(*otherPod) && otherPod.Equal(*pod)
}

func podsNotEqual(pod, otherPod *Pod) bool {
	return !pod.Equal(*otherPod) && !otherPod.Equal(*pod)
}
