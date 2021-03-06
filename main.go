package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"

	"github.com/apex/gateway/v2"
)

func main() {
	server := newServer()
	defer server.Close()

	if isLocal() {
		log.Fatal(http.ListenAndServe(server.port, server))
	} else {
		log.Fatal(gateway.ListenAndServe(server.port, server))
	}
}

func isLocal() bool {
	return os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == ""
}
