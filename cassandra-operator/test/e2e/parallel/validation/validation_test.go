package validation

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test/e2e"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "E2E Suite (Validation Tests)", test.CreateParallelReporters("e2e_validation"))
}

var _ = Context("Cassandra resource validation", func() {
	// The defaulting webhook is called before the OpenAPI validation occurs,
	// See https://kubernetes.io/blog/2019/03/21/a-guide-to-kubernetes-admission-controllers/
	// So the invalid manifests that we attempt to apply here, will instead be rejected by the defaulting webhook,
	// because it will fail to Unmarshal the content.
	// This is fixed in Kubernetes 1.15, which introduces a structural schema  validation step before the defaulting webhook is called.
	// See: https://kubernetes.io/blog/2019/06/20/crd-structural-schema/
	// Skip these tests for now.
	XContext("openapi validation", func() {
		It("should fail with an incomplete cassandra spec", func() {
			command, output, err := e2e.Kubectl(e2e.Namespace, "apply", "-f", "testdata/incomplete-spec.yaml")
			Expect(err).To(HaveOccurred(), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
			Expect(string(output)).To(ContainSubstring(`spec.racks in body is required`))
		})

		It("should fail with an invalid field in a cassandra spec", func() {
			command, output, err := e2e.Kubectl(e2e.Namespace, "apply", "-f", "testdata/invalid-value-spec.yaml")
			Expect(err).To(HaveOccurred(), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
			Expect(string(output)).To(ContainSubstring(`spec.racks.replicas in body must be of type integer: "string"`))
		})
	})

	Context("webhook validation", func() {
		It("should fail if cassandra spec has no racks", func() {
			command, output, err := e2e.Kubectl(
				e2e.Namespace, "apply", "-f", "testdata/spec-no-racks.yaml",
			)
			Expect(err).To(
				HaveOccurred(),
				fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)),
			)
			Expect(string(output)).To(
				ContainSubstring(`admission webhook "vcass.core.sky.uk" denied the request`),
			)
		})
	})
})
