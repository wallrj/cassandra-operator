package validation

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/test/e2e"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "E2E Suite (Validation Tests)", test.CreateSequentialReporters("e2e_validation"))
}

// The defaulting webhook is called before the OpenAPI validation occurs,
// See https://kubernetes.io/blog/2019/03/21/a-guide-to-kubernetes-admission-controllers/
// So the invalid manifests that we attempt to apply here, will instead be rejected by the defaulting webhook,
// because it will fail to Unmarshal the content.
// This is fixed in Kubernetes 1.15, which introduces a structural schema  validation step before the defaulting webhook is called.
// See: https://kubernetes.io/blog/2019/06/20/crd-structural-schema/
// To work around this, we temporarily reconfigure the deployed webhook, giving it a non-matching namespaceSelector.
var _ = e2e.SequentialTestBeforeSuite(func() {
	command, output, err := e2e.Kubectl(
		"",
		"patch",
		"validatingwebhookconfigurations.admissionregistration.k8s.io",
		"validating-webhook-configuration",
		"--type=json",
		"--patch",
		`[{"op": "replace", "path": "/webhooks/0/namespaceSelector", "value": {"matchExpressions": [{"key": "admission.cassandras.core.sky.uk/enabled", "operator": "Exists"}]}}]`,
	)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
	Eventually(func() error {
		command, output, err := e2e.Kubectl(
			e2e.Namespace,
			"apply", "-f", "testdata/spec-no-racks.yaml",
		)
		return errors.Wrapf(err, "command %v failed with output %q", command, string(output))
	}).Should(Succeed())
})

var _ = AfterSuite(func() {
	command, output, err := e2e.Kubectl(
		"",
		"patch",
		"validatingwebhookconfigurations.admissionregistration.k8s.io",
		"validating-webhook-configuration",
		"--type=json",
		"--patch",
		`[{"op": "remove", "path": "/webhooks/0/namespaceSelector"}]`,
	)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Command was: %v \nOutput was %v", command, string(output)))
	Eventually(func() error {
		command, output, err := e2e.Kubectl(
			e2e.Namespace,
			"apply", "-f", "testdata/spec-no-racks.yaml",
		)
		return errors.Wrapf(err, "command %v failed with output %q", command, string(output))
	}).Should(Not(Succeed()))
})

var _ = Context("Cassandra resource validation", func() {
	Context("openapi validation", func() {
		It("should fail with an incomplete cassandra spec", func() {
			Eventually(func() error {
				command, output, err := e2e.Kubectl(
					e2e.Namespace,
					"apply", "-f", "testdata/incomplete-spec.yaml",
				)
				return errors.Wrapf(err, "command %v failed with output %q", command, string(output))
			}).Should(MatchError(ContainSubstring(`spec.racks in body is required`)))
		})

		It("should fail with an invalid field in a cassandra spec", func() {
			Eventually(func() error {
				command, output, err := e2e.Kubectl(
					e2e.Namespace,
					"apply", "-f", "testdata/invalid-value-spec.yaml",
				)
				return errors.Wrapf(err, "command %v failed with output %q", command, string(output))
			}).Should(MatchError(ContainSubstring(`spec.racks.replicas in body must be of type integer: "string"`)))
		})
	})
})
