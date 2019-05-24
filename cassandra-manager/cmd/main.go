package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/metrics"
	"k8s.io/apimachinery/pkg/api/resource"
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

	cluster, err := cluster.New(
		&v1alpha1.Cassandra{
			ObjectMeta: metav1.ObjectMeta{Name: clusterName, Namespace: clusterNamespace},
			Spec: v1alpha1.CassandraSpec{
				Racks: []v1alpha1.Rack{{Name: "a", Replicas: 1, StorageClass: "some-storage", Zone: "some-zone"}},
				Pod: v1alpha1.Pod{
					Memory:      resource.MustParse("1Gi"),
					CPU:         resource.MustParse("100m"),
					StorageSize: resource.MustParse("1Gi"),
				},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	nt := metrics.NewNodetool(cluster, nil)

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		log.Println("/ready called")
		ready, err := nt.IsNodeReady(nodeAddress)
		if err != nil {
			log.Println(err)
		}
		if !ready {
			w.WriteHeader(500)
			log.Println("IsLocalNodeReady failed")
		}
		fmt.Fprintf(w, "cassandra node status: ready: %#v", ready)
	})

	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		log.Println("/live called")
		/*
			if !cql.IsLocalNodeResponding() {
				w.WriteHeader(500)
				log.Println("IsLocalNodeResponding failed")
			}
		*/
		fmt.Fprintf(w, "cassandra node status: ready")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
