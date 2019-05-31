package metrics

import (
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/test"
)

func TestMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Metrics Suite", test.CreateParallelReporters("metrics"))
}

var (
	server    *httptest.Server
	serverURL string
	jolokia   *jolokiaHandler
)

var _ = BeforeSuite(func() {
	jolokia = &jolokiaHandler{}
	server = httptest.NewServer(jolokia)
	serverURL = server.URL
})

var _ = AfterSuite(func() {
	server.Close()
})
