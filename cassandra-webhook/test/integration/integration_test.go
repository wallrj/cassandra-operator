package integration_test

import (
	"io/ioutil"
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
	env = &envtest.Environment{
		CRDDirectoryPaths: []string{"../../../cassandra-operator/kubernetes-resources"},
	}
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
		stdoutBuffer, stderrBuffer, err := env.ControlPlane.KubeCtl().Run("get", "crd")
		Expect(err).To(Not(HaveOccurred()))

		stdout, err := ioutil.ReadAll(stdoutBuffer)
		Expect(err).To(Not(HaveOccurred()))

		stderr, err := ioutil.ReadAll(stderrBuffer)
		Expect(err).To(Not(HaveOccurred()))

		Expect(string(stdout)).To(Equal(""))
		Expect(string(stderr)).To(Equal(""))
	})
})
