package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sky-uk/cassandra-operator/cassandra-manager/pkg/nodetool"
)

func main() {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if !nodetool.IsLocalNodeReady() {
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
