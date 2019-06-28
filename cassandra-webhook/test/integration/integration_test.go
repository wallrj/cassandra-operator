package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// wget https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-1.14.1-linux-amd64.tar.gz
// tar xf kubebuilder-tools-1.14.1-linux-amd64.tar.gz
// KUBEBUILDER_ASSETS=$PWD/kubebuilder/bin ginkgo --debug -v ./test/integration/...

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Integration Suite", []Reporter{envtest.NewlineReporter{}})
}

var env *envtest.Environment

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))
	env = &envtest.Environment{}
	_, err := env.Start()
	Expect(err).NotTo(HaveOccurred())

	close(done)
}, envtest.StartTimeout)

var _ = AfterSuite(func(done Done) {
	Expect(env.Stop()).NotTo(HaveOccurred())

	close(done)
}, envtest.StopTimeout)

var _ = Context("Validating Webhook", func() {
	It("should...", func() {
		Expect(nil).To(BeNil())
	})
})
