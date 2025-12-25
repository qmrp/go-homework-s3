package main

import (
	"log"
	"net/http"
)

func main() {
	if err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})); err != nil {
		log.Fatalf("failed to start http server: %v", err)
	}
}
