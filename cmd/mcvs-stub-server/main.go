package main

import (
	"net/http"

	"github.com/schubergphilis/mcvs-integrationtest-services/internal/pkg/constants"
	"github.com/schubergphilis/mcvs-integrationtest-services/internal/stubserver"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Create a new server
	server := stubserver.NewServer()

	// Start server
	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           server.Router,
		ReadHeaderTimeout: constants.DefaultHTTPTimeout,
	}

	log.Info("Starting server on :8080")
	err := httpServer.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server closed")
}
