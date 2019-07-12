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
	RunSpecsWithDefaultAndCustomReporters(t, "E2E Suite (Validation Tests)", test.CreateSequentialReporters("e2e_validation"))
}

var _ = e2e.SequentialTestBeforeSuite(func() {})

var _ = Context("webhook", func() {
	Context("validation", func() {
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
