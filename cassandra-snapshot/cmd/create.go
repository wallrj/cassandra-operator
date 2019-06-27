package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-snapshot/pkg/snapshot"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates snapshots of a cassandra cluster for one or more keyspaces",
	Run:   createSnapshot,
}

var snapshotTimeout time.Duration

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().DurationVarP(&snapshotTimeout, "snapshot-timeout", "t", v1alpha1.DefaultSnapshotTimeoutSeconds*time.Second, "Max wait time for the snapshot creation")
}

func createSnapshot(_ *cobra.Command, _ []string) {
	err := snapshot.New().DoCreate(&snapshot.CreateConfig{
		Keyspaces:       keyspaces,
		PodLabel:        podLabel,
		Namespace:       namespace,
		SnapshotTimeout: snapshotTimeout,
	})

	if err != nil {
		log.Errorf("Error while creating snapshot for pods with labels %s: %v ", podLabel, err)
		os.Exit(1)
	}
}
