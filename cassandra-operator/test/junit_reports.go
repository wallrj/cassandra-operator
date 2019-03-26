package test

import (
	"flag"
	"fmt"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
)

var junitReportDir string

func init() {
	flag.StringVar(&junitReportDir, "junit-report-dir", "", "path to the directory that will contain the test reports")
}

// CreateParallelReporters creates test reporters for parallel tests
func CreateParallelReporters(name string) []ginkgo.Reporter {
	return createReporters(fmt.Sprintf("%s_%d.xml", name, config.GinkgoConfig.ParallelNode))
}

// CreateSequentialReporters creates test reporters for sequential tests
func CreateSequentialReporters(name string) []ginkgo.Reporter {
	return createReporters(fmt.Sprintf("%s.xml", name))
}

func createReporters(filename string) []ginkgo.Reporter {
	if junitReportDir == "" {
		return []ginkgo.Reporter{}
	}

	config.DefaultReporterConfig.Verbose = true
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("%s/%s", junitReportDir, filename))
	return []ginkgo.Reporter{junitReporter}
}
