package provisioning

import (
	"context"
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"os"
)

const project = "organization-provisioner"

type Result struct {
	Realm string
}

func Provisioner(ctx context.Context, orgId string, orgName string) (*Result, error) {
	stackName := orgId

	program := deployer(orgId, orgName)

	workDirPath := fmt.Sprintf("/tmp/%s", stackName)

	err := os.MkdirAll(workDirPath, os.ModePerm)

	if err != nil {
		return nil, err
	}

	workDir := auto.WorkDir(workDirPath)

	s, err := auto.UpsertStackInlineSource(ctx, stackName, project, program, workDir)

	if err != nil {
		return nil, err
	}

	upResult, err := s.Up(ctx, optup.ProgressStreams(os.Stdout))

	if err != nil {
		return nil, err
	}

	return &Result{
		Realm: upResult.Outputs["realm"].Value.(string),
	}, nil
}
