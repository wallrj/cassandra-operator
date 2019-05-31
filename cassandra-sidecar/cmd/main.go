package main

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	logLevel         string
	nodeAddress      string
	clusterName      string
	clusterNamespace string
)

var rootCmd = &cobra.Command{
	Use:               "cassandra-sidecar",
	Short:             "Sidecar for interacting with Cassandra nodes",
	PersistentPreRunE: handleArgs,
	RunE:              start,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Execute failed")
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", log.InfoLevel.String(), "should be one of: debug, info, warn, error, fatal, panic")
}

func handleArgs(_ *cobra.Command, _ []string) error {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log-level")
	}
	log.SetLevel(level)

	nodeAddress = os.Getenv("NODE_LISTEN_ADDRESS")
	if nodeAddress == "" {
		return fmt.Errorf("NODE_LISTEN_ADDRESS must be set")
	}

	clusterName = os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		return fmt.Errorf("CLUSTER_NAME must be set")
	}

	clusterNamespace = os.Getenv("CLUSTER_NAMESPACE")
	if clusterNamespace == "" {
		return fmt.Errorf("CLUSTER_NAMESPACE must be set")
	}

	return nil
}

func start(_ *cobra.Command, _ []string) error {
	logger := log.WithFields(
		log.Fields{
			"clusterName":      clusterName,
			"clusterNamespace": clusterNamespace,
			"nodeAddress":      nodeAddress,
		},
	)
	nt := metrics.NewNodetool(
		cluster.NewWithoutValidation(
			&v1alpha1.Cassandra{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterName,
					Namespace: clusterNamespace,
				},
			},
		),
		nil,
	)

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ready, err := nt.IsNodeReady(nodeAddress)
		if err != nil {
			w.WriteHeader(500)
			logger.WithError(err).Error("IsNodeReady failed")
			return
		}
		if !ready {
			w.WriteHeader(500)
			return
		}
		_, err = fmt.Fprintf(w, "ok")
		if err != nil {
			log.WithError(err).Error("Fprintf failed")
		}
	})

	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		_, err := nt.IsNodeReady(nodeAddress)
		if err != nil {
			w.WriteHeader(500)
			logger.WithError(err).Error("IsNodeReady failed")
			return
		}
		_, err = fmt.Fprintf(w, "ok")
		if err != nil {
			log.WithError(err).Error("Fprintf failed")
		}
	})

	return http.ListenAndServe(":8080", nil)
}
