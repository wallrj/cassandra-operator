package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "cassandra-snapshot",
	Short: "Creates or cleans up snapshots of a cassandra cluster for one or more keyspaces",
	Args:  cobra.MinimumNArgs(1),
}

var (
	keyspaces []string
	logLevel  string
	podLabel  string
	namespace string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringSliceVarP(&keyspaces, "keyspace", "k", []string{}, "Keyspace to snapshot. Repeat this flag to specify multiple values.")
	rootCmd.PersistentFlags().StringVarP(&podLabel, "pod-label", "l", "", "Kubernetes labels attached to cassandra pods that are targeted. Comma-separated list")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace where the cassandra pods are deployed")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", log.InfoLevel.String(), "should be one of: debug, info, warn, error, fatal, panic")
	rootCmd.MarkFlagRequired("pod-label")
	rootCmd.MarkFlagRequired("namespace")
	cobra.OnInitialize(onInitialise)
}

func onInitialise() {
	setLogLevel(logLevel)
}

func setLogLevel(logLevel string) {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		logAndExit("invalid log-level")
	}
	log.SetLevel(level)
}

func logAndExit(message string, args ...interface{}) {
	log.Errorf(message, args...)
	os.Exit(1)
}
