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
	Context("openapi validation", func() {
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
		var namespace string

		BeforeEach(func() {
			var err error
			namespace, err = e2e.CreateNamespace()
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := e2e.DeleteNamespace(namespace)
			Expect(err).To(Not(HaveOccurred()))
		})
		Context("disabled", func() {
			It("should succeed if cassandra spec has no racks", func() {
				command, output, err := e2e.Kubectl(
					namespace, "apply", "-f", "testdata/spec-no-racks.yaml",
				)
				Expect(err).ToNot(
					HaveOccurred(),
					fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)),
				)
			})
		})

		Context("enabled", func() {
			BeforeEach(func() {
				command, output, err := e2e.Kubectl(
					"", "label", "--overwrite", "namespace",
					namespace, "webhooks.cassandras.core.sky.uk=enabled",
				)
				Expect(err).ToNot(
					HaveOccurred(),
					fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)),
				)
			})
			It("should fail if cassandra spec has no racks", func() {
				command, output, err := e2e.Kubectl(
					namespace, "apply", "-f", "testdata/spec-no-racks.yaml",
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
})
