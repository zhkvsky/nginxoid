// Package main implements a mock HTTP server serving sample NGINX logs for local testing.
package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	port := flag.String("port", "8000", "server port")

	flag.Parse()

	address := net.JoinHostPort("127.0.0.1", *port)
	server := &http.Server{
		Addr:              address,
		Handler:           http.FileServer(http.Dir("sample_logs")),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("sample logs: http://%s/access.txt", address)
	log.Fatal(server.ListenAndServe())
}
