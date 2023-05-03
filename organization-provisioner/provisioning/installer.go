package provisioning

import (
	"context"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// Installer ensures plugins runs once before the server boots up
// making sure the proper pulumi plugins are installed
func Installer(ctx context.Context) error {
	w, err := auto.NewLocalWorkspace(ctx)

	if err != nil {
		return err
	}

	err = w.InstallPlugin(ctx, "keycloak", "v5.1.0")

	if err != nil {
		return err
	}

	return nil
}
