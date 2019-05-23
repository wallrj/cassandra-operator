package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "cassandra node status: ready")
	})

	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "cassandra node status: ready")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
