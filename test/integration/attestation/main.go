/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	serveCertPath := os.Getenv("TLS_CERT_PATH")
	serveKeyPath := os.Getenv("TLS_KEY_PATH")

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		log.Fatalf("LISTEN_ADDR is required")

		return
	}
	log.Printf("Listening on %s", listenAddr)

	srv := newServer(&serverConfig{})

	if serveCertPath == "" || serveKeyPath == "" {
		log.Fatal(http.ListenAndServe(
			listenAddr,
			srv,
		))

		return
	}

	log.Fatal(http.ListenAndServeTLS(
		listenAddr,
		serveCertPath, serveKeyPath,
		srv,
	))
}
