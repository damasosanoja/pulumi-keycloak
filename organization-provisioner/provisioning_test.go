package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"net/http"
	"testing"
)

type keycloakContainer struct {
	testcontainers.Container
	BaseUrl            string
	AccessToken        string
	Client             *gocloak.GoCloak
	PulumiClientSecret string
	ContainerBaseUrl   string
}

func setupKeycloak(ctx context.Context) (*keycloakContainer, error) {
	kcReq := testcontainers.ContainerRequest{
		Image:        "quay.io/keycloak/keycloak:21.0.2",
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"KEYCLOAK_ADMIN":          "admin",
			"KEYCLOAK_ADMIN_PASSWORD": "admin",
		},
		Cmd:        []string{"start-dev"},
		WaitingFor: wait.ForLog("Running the server in development"),
	}

	kcC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kcReq,
		Started:          true,
	})

	if err != nil {
		return nil, err
	}

	kcHost, err := kcC.Host(ctx)
	kcPort, err := kcC.MappedPort(ctx, "8080")

	kcUrl := fmt.Sprintf("http://%s:%s", kcHost, kcPort.Port())

	fmt.Println(kcUrl)

	kcClient := gocloak.NewClient(kcUrl)

	token, err := kcClient.LoginAdmin(ctx, "admin", "admin", "master")

	if err != nil {
		return nil, err
	}

	clientReq := gocloak.Client{
		ClientID:               gocloak.StringP("pulumi"),
		ServiceAccountsEnabled: gocloak.BoolP(true),
		Enabled:                gocloak.BoolP(true),
		Protocol:               gocloak.StringP("openid-connect"),
		PublicClient:           gocloak.BoolP(false),
	}

	cId, err := kcClient.CreateClient(ctx, token.AccessToken, "master", clientReq)

	if err != nil {
		return nil, err
	}

	sa, err := kcClient.GetClientServiceAccount(ctx, token.AccessToken, "master", cId)

	if err != nil {
		return nil, err
	}

	role, err := kcClient.GetRealmRole(ctx, token.AccessToken, "master", "admin")

	if err != nil {
		return nil, err
	}

	err = kcClient.AddRealmRoleToUser(ctx, token.AccessToken, "master", *sa.ID, []gocloak.Role{*role})

	if err != nil {
		return nil, err
	}

	pulumiClient, err := kcClient.GetClient(ctx, token.AccessToken, "master", cId)

	if err != nil {
		return nil, err
	}

	fmt.Println(*pulumiClient.ClientID, *pulumiClient.Secret)

	cIp, err := kcC.ContainerIP(ctx)
	if err != nil {
		return nil, err
	}

	containerBaseUrl := fmt.Sprintf("http://%s:%s", cIp, "8080")

	return &keycloakContainer{
		Container:          kcC,
		AccessToken:        token.AccessToken,
		BaseUrl:            kcUrl,
		Client:             kcClient,
		PulumiClientSecret: *pulumiClient.Secret,
		ContainerBaseUrl:   containerBaseUrl,
	}, nil
}

type orgProvContainer struct {
	testcontainers.Container
	BaseUrl string
}

func setupOrgProvContainer(ctx context.Context, kcC *keycloakContainer) (*orgProvContainer, error) {
	provReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    ".",
			Dockerfile: "Dockerfile",
		},
		ExposedPorts: []string{"8000/tcp"},
		Env: map[string]string{
			"PULUMI_BACKEND_URL":       "file:///tmp/pulumi",
			"PULUMI_CONFIG_PASSPHRASE": "some-passphrase",
			"KEYCLOAK_URL":             kcC.ContainerBaseUrl,
			"KEYCLOAK_CLIENT_ID":       "pulumi",
			"KEYCLOAK_CLIENT_SECRET":   kcC.PulumiClientSecret,
		},
		WaitingFor: wait.ForLog("8000"),
	}

	provC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: provReq,
		Started:          true,
	})

	if err != nil {
		return nil, err
	}

	host, err := provC.Host(ctx)
	port, err := provC.MappedPort(ctx, "8000")

	baseUrl := fmt.Sprintf("http://%s:%s", host, port.Port())

	fmt.Println(baseUrl)

	return &orgProvContainer{
		Container: provC,
		BaseUrl:   baseUrl,
	}, nil
}

func TestOrganizationProvisioner(t *testing.T) {
	ctx := context.Background()

	kcC, err := setupKeycloak(ctx)

	if err != nil {
		t.Error(err)
		return
	}

	orgProvC, err := setupOrgProvContainer(ctx, kcC)

	if err != nil {
		t.Error(err)
		return
	}

	req, err := json.Marshal(map[string]string{
		"id":   "zone2",
		"name": "Zone 2 tech",
	})

	endpoint := fmt.Sprintf("%s/provisioner/organizations", orgProvC.BaseUrl)

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(req))

	if err != nil {
		t.Error(err)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()

		if err != nil {
			t.Error(err)
			return
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(body))

	defer func() {
		if err := kcC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err.Error())
		}
		if err := orgProvC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err.Error())
		}
	}()
}
