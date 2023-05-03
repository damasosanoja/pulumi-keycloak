package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/pulumi/automation-api-examples/go/pulumi_over_http/provisioning"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	err := provisioning.Installer(ctx)

	if err != nil {
		log.Fatal("Failed to install provisioning dependencies")
	}

	router := mux.NewRouter()

	router.HandleFunc("/provisioner/organizations", provisionOrganization).Methods("POST")
	router.HandleFunc("/provisioner/organizations/{id}", deprovisionOrganization).Methods("DELETE")

	s := &http.Server{Addr: ":8000", Handler: router}

	err = s.ListenAndServe()

	if err != nil {
		log.Fatal("Unable to start server", err)
	}
}
