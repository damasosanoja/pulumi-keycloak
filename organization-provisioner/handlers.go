package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pulumi/automation-api-examples/go/pulumi_over_http/provisioning"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"net/http"
)

type ProvisionOrganizationRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProvisioningOrganizationResponse struct {
	Realm string `json:"realm"`
}

func provisionOrganization(w http.ResponseWriter, req *http.Request) {
	var provisionReq ProvisionOrganizationRequest
	err := json.NewDecoder(req.Body).Decode(&provisionReq)

	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)

		return
	}

	if len(provisionReq.ID) == 0 {
		http.Error(w, "id is required", http.StatusBadRequest)

		return
	}

	if len(provisionReq.Name) == 0 {
		http.Error(w, "name is required", http.StatusBadRequest)

		return
	}

	ctx := req.Context()

	result, err := provisioning.Provisioner(ctx, provisionReq.ID, provisionReq.Name)

	if err != nil {
		if auto.IsConcurrentUpdateError(err) {
			http.Error(w, "Update is already in progress", http.StatusConflict)

			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	response := &ProvisioningOrganizationResponse{
		Realm: result.Realm,
	}

	err = json.NewEncoder(w).Encode(&response)

	if err != nil {
		http.Error(w, "Could not write response json", http.StatusInternalServerError)
	}
}

func deprovisionOrganization(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	params := mux.Vars(req)

	if len(params["id"]) == 0 {
		http.Error(w, "id is required", http.StatusBadRequest)

		return
	}

	err := provisioning.Deprovisioner(ctx, params["id"])

	if err != nil {
		if auto.IsConcurrentUpdateError(err) {
			http.Error(w, "Deletion is already in progress", http.StatusConflict)

			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
