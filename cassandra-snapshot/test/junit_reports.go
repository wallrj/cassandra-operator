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

// CreateReporters creates test reporters for all tests
func CreateReporters(name string) []ginkgo.Reporter {
	filename := fmt.Sprintf("%s.xml", name)
	if junitReportDir == "" {
		return []ginkgo.Reporter{}
	}

	config.DefaultReporterConfig.Verbose = true
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("%s/%s", junitReportDir, filename))
	return []ginkgo.Reporter{junitReporter}
}
