package cmd

import (
	"github.com/sky-uk/cassandra-operator/cassandra-snapshot/test"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const snapshotCommand = "../../../build/bin/cassandra-snapshot"

func TestCommandLine(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Command Line Suite", test.CreateReporters("commandline"))
}

var _ = Describe("cassandra-snapshot command line", func() {
	Describe("--help", func() {
		It("should print available flags", func() {
			output, err := exec.Command(snapshotCommand, "--help").CombinedOutput()
			Expect(err).ToNot(HaveOccurred(), string(output))
			Expect(string(output)).To(ContainSubstring("create      Creates snapshots of a cassandra cluster for one or more keyspaces"))
			Expect(string(output)).To(ContainSubstring("cleanup     Removes snapshots of a cassandra cluster older than the retention period"))
			Expect(string(output)).To(ContainSubstring("-h, --help"))
			Expect(string(output)).To(ContainSubstring("-L, --log-level"))
		})
	})
	Describe("create --help", func() {
		It("should print available flags", func() {
			output, err := exec.Command(snapshotCommand, "create", "--help").CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("Creates snapshots of a cassandra cluster for one or more keyspaces"))
			Expect(string(output)).To(ContainSubstring("-h, --help"))
			Expect(string(output)).To(ContainSubstring("-k, --keyspace"))
			Expect(string(output)).To(ContainSubstring("-l, --pod-label"))
			Expect(string(output)).To(ContainSubstring("-n, --namespace"))
			Expect(string(output)).To(ContainSubstring("-t, --snapshot-timeout"))
			Expect(string(output)).To(ContainSubstring("-L, --log-level"))
		})
	})
	Describe("cleanup --help", func() {
		It("should print available flags", func() {
			output, err := exec.Command(snapshotCommand, "cleanup", "--help").CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("Removes snapshots of a cassandra cluster older than the retention period"))
			Expect(string(output)).To(ContainSubstring("-h, --help"))
			Expect(string(output)).To(ContainSubstring("-k, --keyspace"))
			Expect(string(output)).To(ContainSubstring("-l, --pod-label"))
			Expect(string(output)).To(ContainSubstring("-n, --namespace"))
			Expect(string(output)).To(ContainSubstring("-r, --retention-period"))
			Expect(string(output)).To(ContainSubstring("-t, --cleanup-timeout"))
			Expect(string(output)).To(ContainSubstring("-L, --log-level"))
		})
	})
})
