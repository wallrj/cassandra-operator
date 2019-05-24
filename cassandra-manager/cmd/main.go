package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
)

func main() {

	cluster := cluster.New(&v1alpha1.Cassandra{})
	nt := NewNodetool(nil, nil)

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if !nt.IsNodeReady("192.0.2.1") {
			w.WriteHeader(500)
			log.Println("IsLocalNodeReady failed")
		}
		fmt.Fprintf(w, "cassandra node status: ready")
	})

	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
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
