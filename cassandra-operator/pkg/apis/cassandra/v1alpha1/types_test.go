package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
)

func TestTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Types Suite", test.CreateParallelReporters("types"))
}

var _ = Describe("Cassandra Types", func() {
	Context("Pod", func() {

		DescribeTable("equality",
			equalityCheck,
			Entry("if all fields are equal", func(pod *Pod) {}),
			Entry("when cpu value is the same but using a different amount", func(pod *Pod) { pod.CPU = resource.MustParse("1") }),
			Entry("when memory value is the same but using a different amount", func(pod *Pod) { pod.Memory = resource.MustParse("2048Mi") }),
			Entry("when storage size value is the same but using a different amount", func(pod *Pod) { pod.StorageSize = resource.MustParse("1024Mi") }),
		)

		DescribeTable("inequality",
			inEqualityCheck,
			Entry("when one pod has a nil bootstrap image", func(pod *Pod) { pod.BootstrapperImage = nil }),
			Entry("when pods have different bootstrap images", func(pod *Pod) { pod.BootstrapperImage = ptr.String("another image") }),
			Entry("when one pod has a nil sidecar image", func(pod *Pod) { pod.SidecarImage = nil }),
			Entry("when pods have different sidecar images", func(pod *Pod) { pod.SidecarImage = ptr.String("another image") }),
			Entry("when one pod has a nil cassandra image", func(pod *Pod) { pod.Image = nil }),
			Entry("when pods have different cassandra images", func(pod *Pod) { pod.Image = ptr.String("another image") }),
			Entry("when one pod has no storage size", func(pod *Pod) { pod.StorageSize = resource.Quantity{} }),
			Entry("when pods have different storage sizes", func(pod *Pod) { pod.StorageSize = resource.MustParse("10Gi") }),
			Entry("when one pod has no cpu", func(pod *Pod) { pod.CPU = resource.Quantity{} }),
			Entry("when pods have different number of cpu", func(pod *Pod) { pod.CPU = resource.MustParse("10") }),
			Entry("when one pod has no memory", func(pod *Pod) { pod.Memory = resource.Quantity{} }),
			Entry("when pods have different memory sizes", func(pod *Pod) { pod.Memory = resource.MustParse("10") }),
			Entry("when liveness probes have different success threshold values", func(pod *Pod) { pod.LivenessProbe.SuccessThreshold = ptr.Int32(20) }),
			Entry("when liveness probes have different timeout values", func(pod *Pod) { pod.LivenessProbe.TimeoutSeconds = ptr.Int32(20) }),
			Entry("when liveness probes have different failure threshold values", func(pod *Pod) { pod.LivenessProbe.FailureThreshold = ptr.Int32(20) }),
			Entry("when liveness probes have different initial delay values", func(pod *Pod) { pod.LivenessProbe.InitialDelaySeconds = ptr.Int32(20) }),
			Entry("when liveness probes have different period seconds values", func(pod *Pod) { pod.LivenessProbe.PeriodSeconds = ptr.Int32(20) }),
			Entry("when one liveness probe has a nil success threshold", func(pod *Pod) { pod.LivenessProbe.SuccessThreshold = nil }),
			Entry("when one liveness probe has a nil timeout", func(pod *Pod) { pod.LivenessProbe.TimeoutSeconds = nil }),
			Entry("when one liveness probe has a nil failure threshold", func(pod *Pod) { pod.LivenessProbe.FailureThreshold = nil }),
			Entry("when one liveness probe has a nil delay", func(pod *Pod) { pod.LivenessProbe.InitialDelaySeconds = nil }),
			Entry("when one liveness probe has a nil period", func(pod *Pod) { pod.LivenessProbe.PeriodSeconds = nil }),
			Entry("when readiness probes have different timeout values", func(pod *Pod) { pod.ReadinessProbe.TimeoutSeconds = ptr.Int32(20) }),
			Entry("when readiness probes have different failure threshold values", func(pod *Pod) { pod.ReadinessProbe.FailureThreshold = ptr.Int32(20) }),
			Entry("when readiness probes have different initial delay values", func(pod *Pod) { pod.ReadinessProbe.InitialDelaySeconds = ptr.Int32(20) }),
			Entry("when readiness probes have different period seconds values", func(pod *Pod) { pod.ReadinessProbe.PeriodSeconds = ptr.Int32(20) }),
			Entry("when one readiness probe has a nil success threshold", func(pod *Pod) { pod.ReadinessProbe.SuccessThreshold = nil }),
			Entry("when one readiness probe has a nil timeout", func(pod *Pod) { pod.ReadinessProbe.TimeoutSeconds = nil }),
			Entry("when one readiness probe has a nil failure threshold", func(pod *Pod) { pod.ReadinessProbe.FailureThreshold = nil }),
			Entry("when one readiness probe has a nil delay", func(pod *Pod) { pod.ReadinessProbe.InitialDelaySeconds = nil }),
			Entry("when one readiness probe has a nil period", func(pod *Pod) { pod.ReadinessProbe.PeriodSeconds = nil }),
		)
	})
})

func equalityCheck(applyChange func(pod *Pod)) {
	comparisonCheck(applyChange, podsEqual)
}

func inEqualityCheck(applyChange func(pod *Pod)) {
	comparisonCheck(applyChange, podsNotEqual)
}

func comparisonCheck(applyChange func(pod *Pod), expectCheck func(pod, otherPod *Pod) bool) {
	pod1 := &Pod{
		BootstrapperImage: ptr.String("BootstrapperImage"),
		SidecarImage:      ptr.String("SidecarImage"),
		Image:             ptr.String("Image"),
		Memory:            resource.MustParse("2Gi"),
		CPU:               resource.MustParse("1000m"),
		StorageSize:       resource.MustParse("1Gi"),
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
	pod2 := pod1.DeepCopy()

	applyChange(pod1)

	Expect(expectCheck(pod1, pod2)).To(BeTrue())
}

func podsEqual(pod, otherPod *Pod) bool {
	return pod.Equal(*otherPod) && otherPod.Equal(*pod)
}

func podsNotEqual(pod, otherPod *Pod) bool {
	return !pod.Equal(*otherPod) && !otherPod.Equal(*pod)
}
