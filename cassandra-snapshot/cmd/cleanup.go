package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-snapshot/pkg/snapshot"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Removes snapshots of a cassandra cluster older than the retention period",
	Run:   cleanupSnapshot,
}

var (
	retentionPeriod time.Duration
	cleanupTimeout  time.Duration
)

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().DurationVarP(&retentionPeriod, "retention-period", "r", 7*24*time.Hour, "Duration backups should be kept for")
	cleanupCmd.Flags().DurationVarP(&cleanupTimeout, "cleanup-timeout", "t", 10*time.Second, "Max wait time for the cleanup operation")
}

func cleanupSnapshot(_ *cobra.Command, _ []string) {
	err := snapshot.New().DoCleanup(&snapshot.CleanupConfig{
		Namespace:       namespace,
		RetentionPeriod: retentionPeriod,
		Keyspaces:       keyspaces,
		PodLabel:        podLabel,
		CleanupTimeout:  cleanupTimeout,
	})

	if err != nil {
		log.Errorf("Error while cleaning up snapshots for pods with labels %s: %v ", podLabel, err)
		os.Exit(1)
	}
}
