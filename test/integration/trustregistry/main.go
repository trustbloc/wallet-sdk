package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	rulesFile := os.Getenv("RULES_FILE_PATH")
	if rulesFile == "" {
		log.Fatalf("RULES_FILE_PATH is required")

		return
	}

	serveCertPath := os.Getenv("TLS_CERT_PATH")
	if serveCertPath == "" {
		log.Fatalf("TLS_CERT_PATH is required")

		return
	}

	serveKeyPath := os.Getenv("TLS_KEY_PATH")
	if serveKeyPath == "" {
		log.Fatalf("TLS_KEY_PATH is required")

		return
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		log.Fatalf("LISTEN_ADDR is required")

		return
	}

	server, err := NewServer(&Config{rulesJSONPath: rulesFile})
	if err != nil {
		log.Fatalf("failed to initialize server: %s", err)

		return
	}

	log.Printf("Listening on %s", listenAddr)

	log.Fatal(http.ListenAndServeTLS(
		listenAddr,
		serveCertPath, serveKeyPath,
		server,
	))
}
