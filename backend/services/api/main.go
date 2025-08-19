package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("CORE_API_PORT")
	if port == "" {
		port = "4000"
	}
	addr := ":" + port
	log.Printf("starting api on %s", addr)
	if err := http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
