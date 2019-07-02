package validation

import (
	"fmt"
	"testing"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
	. "github.com/sky-uk/cassandra-operator/cassandra-operator/test/e2e"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "E2E Suite (Validation Tests)", test.CreateParallelReporters("e2e_validation"))
}

var _ = Context("Cassandra resource validation", func() {

	Context("openapi validation", func() {
		It("should fail with an incomplete cassandra spec", func() {
			command, output, err := Kubectl(Namespace, "apply", "-f", "testdata/incomplete-spec.yaml")
			Expect(err).To(HaveOccurred(), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
			Expect(string(output)).To(ContainSubstring(`spec.racks in body is required`))
		})

		It("should fail with an invalid field in a cassandra spec", func() {
			command, output, err := Kubectl(Namespace, "apply", "-f", "testdata/invalid-value-spec.yaml")
			Expect(err).To(HaveOccurred(), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
			Expect(string(output)).To(ContainSubstring(`spec.racks.replicas in body must be of type integer: "string"`))
		})
	})

	Context("webhook validation disabled", func() {
		AfterEach(func() {
			command, output, err := Kubectl(Namespace, "delete", "-f", "testdata/spec-no-racks.yaml")
			Expect(err).To(Not(HaveOccurred()), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
		})

		It("should succeed with a slightly invalid cassandra spec when webhook is not enabled for namespace", func() {
			command, output, err := Kubectl(Namespace, "apply", "-f", "testdata/spec-no-racks.yaml")
			Expect(err).To(Not(HaveOccurred()), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
		})
	})

	Context("webhook validation enabled", func() {
		BeforeEach(func() {
			command, output, err := Kubectl(Namespace, "label", "namespace", Namespace, "--overwrite", "webhooks.cassandra.core.sky.uk=enabled")
			Expect(err).To(Not(HaveOccurred()), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
		})

		AfterEach(func() {
			command, output, err := Kubectl(Namespace, "label", "namespace", Namespace, "webhooks.cassandra.core.sky.uk-")
			Expect(err).To(Not(HaveOccurred()), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
		})

		It("should fail with a slightly invalid cassandra spec when webhook is enabled for namespace", func() {
			command, output, err := Kubectl(Namespace, "apply", "-f", "testdata/spec-no-racks.yaml")
			Expect(err).To(HaveOccurred(), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
			Expect(string(output)).To(ContainSubstring(`admission webhook "vcass.core.sky.uk" denied the request`))
		})
	})
})
