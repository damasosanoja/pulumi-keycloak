package provisioning

import (
	"context"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"os"
)

func Deprovisioner(ctx context.Context, orgId string) error {
	// program doesn't matter for destroying a stack
	program := func(ctx *pulumi.Context) error {
		return nil
	}

	s, err := auto.SelectStackInlineSource(ctx, orgId, project, program)

	if err != nil {
		return err
	}

	// destroy the stack
	// we'll write all the logs to stdout, so we can watch requests get processed
	_, err = s.Destroy(ctx, optdestroy.ProgressStreams(os.Stdout))

	if err != nil {
		return err
	}

	// delete the stack and all associated history and config
	err = s.Workspace().RemoveStack(ctx, orgId)

	if err != nil {
		return err
	}

	return nil
}
