package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"

	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/client/clientset/versioned"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/operator"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	metricPollInterval   time.Duration
	metricRequestTimeout time.Duration
	logLevel             string
	allowEmptyDir        bool
)

var rootCmd = &cobra.Command{
	Use:               "cassandra-operator",
	Short:             "Operator for provisioning Cassandra clusters.",
	PersistentPreRunE: handleArgs,
	RunE:              startOperator,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().DurationVar(&metricPollInterval, "metric-poll-interval", 5*time.Second, "Poll interval between cassandra nodes metrics retrieval")
	rootCmd.PersistentFlags().DurationVar(&metricRequestTimeout, "metric-request-timeout", 2*time.Second, "Time limit for cassandra node metrics requests")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", log.InfoLevel.String(), "should be one of: debug, info, warn, error, fatal, panic")
	// TODO: Remove this --empty-dir command line flag
	rootCmd.PersistentFlags().BoolVar(&allowEmptyDir, "allow-empty-dir", false, "Set to true in order to allow creation of clusters which use emptyDir storage")
}

func handleArgs(_ *cobra.Command, _ []string) error {
	var isPositive = func(duration time.Duration) bool {
		currentTime := time.Now()
		return currentTime.Add(duration).After(currentTime)
	}

	if !isPositive(metricPollInterval) {
		return fmt.Errorf("invalid metric-poll-interval, it must be a positive integer")
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log-level")
	}
	log.SetLevel(level)

	return nil
}

func startOperator(_ *cobra.Command, _ []string) error {
	operatorConfig := &operator.Config{
		MetricRequestDuration: metricPollInterval,
		MetricPollInterval:    metricPollInterval,
		AllowEmptyDir:         allowEmptyDir,
	}
	log.Infof("Starting Cassandra operator with config: %v", operatorConfig)

	if err := v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return fmt.Errorf("unable to register metadata for Cassandra resources: %v", err)
	}

	kubernetesConfig := kubernetesConfig()
	op := operator.New(kubernetesClient(kubernetesConfig), cassandraClient(kubernetesConfig), operatorConfig)
	op.Run()

	return nil
}

func kubernetesConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Unable to obtain in-cluster config: %v", err)
	}
	return config
}

func kubernetesClient(config *rest.Config) *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to obtain clientset: %v", err)
	}
	return clientset
}

func cassandraClient(config *rest.Config) *versioned.Clientset {
	return versioned.NewForConfigOrDie(config)
}
