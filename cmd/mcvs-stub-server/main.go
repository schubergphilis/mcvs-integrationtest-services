package main

import (
	"net/http"

	"github.com/schubergphilis/mcvs-integrationtest-services/internal/app/stubserver"
	"github.com/schubergphilis/mcvs-integrationtest-services/internal/pkg/constants"
	log "github.com/sirupsen/logrus"
)

func main() {
	server := stubserver.NewServer()

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
