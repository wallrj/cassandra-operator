package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	nodeAddress := os.Getenv("NODE_LISTEN_ADDRESS")
	if nodeAddress == "" {
		log.Fatal("NODE_LISTEN_ADDRESS must be set")
	}

	clusterName := os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		log.Fatal("CLUSTER_NAME must be set")
	}

	clusterNamespace := os.Getenv("CLUSTER_NAMESPACE")
	if clusterNamespace == "" {
		log.Fatal("CLUSTER_NAMESPACE must be set")
	}

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
		log.Println("/ready called")
		ready, err := nt.IsNodeReady(nodeAddress)
		if err != nil {
			log.Println(err)
		}
		if !ready {
			w.WriteHeader(500)
			log.Println("IsNodeReady failed")
		}
		fmt.Fprintf(w, "cassandra node status: ready: %#v", ready)
	})

	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		log.Println("/live called")

		status := "succeeded"

		_, err := nt.IsNodeReady(nodeAddress)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			log.Println("IsNodeReady failed")
			status = "failed"
		}
		fmt.Fprintf(w, "cassandra node status: %s", status)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
