package validation

import (
	"os/exec"
	"testing"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "E2E Suite (Validation Tests)", test.CreateParallelReporters("e2e_validation"))
}

var _ = Context("Cassandra resource validation", func() {
	It("should fail with an incomplete cassandra spec", func() {
		output, err := exec.Command("kubectl", "apply", "-f", "testdata/incomplete-spec.yaml").CombinedOutput()
		Expect(err).To(HaveOccurred())
		Expect(string(output)).To(ContainSubstring(`spec.racks in body is required`))
	})

	It("should fail with an invalid field in a cassandra spec", func() {
		output, err := exec.Command("kubectl", "apply", "-f", "testdata/invalid-value-spec.yaml").CombinedOutput()
		Expect(err).To(HaveOccurred())
		Expect(string(output)).To(ContainSubstring(`spec.racks.replicas in body must be of type integer: "string"`))
	})
})
