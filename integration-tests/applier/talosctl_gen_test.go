package applier_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"github.com/stretchr/testify/assert"
)

// ProxyMock intercepts local.Command resources and executes them for real.
type ProxyMock struct {
	pulumi.MockResourceMonitor
	lastStdout string
}

func (m *ProxyMock) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func (m *ProxyMock) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	if args.TypeToken == "command:local:Command" {
		cmdStr := args.Inputs["create"].StringValue()
		dir := ""
		if args.Inputs["dir"].HasValue() {
			dir = args.Inputs["dir"].StringValue()
		}

		if dir != "" {
			_ = os.MkdirAll(dir, 0o700)
		}

		cmd := exec.Command("sh", "-c", cmdStr)
		if dir != "" {
			cmd.Dir = dir
		}

		output, err := cmd.CombinedOutput()
		m.lastStdout = string(output)
		if err != nil {
			return "", nil, err
		}

		return args.Name + "_id", resource.PropertyMap{
			"stdout": resource.NewStringProperty(m.lastStdout),
			"stderr": resource.NewStringProperty(""),
		}, nil
	}

	return args.Name + "_id", args.Inputs, nil
}

func TestGenerateSecretsWithRealTalosctl(t *testing.T) {
	t.Setenv("PULUMI_MOCK_RESOURCES", "1")

	if _, err := exec.LookPath("talosctl"); err != nil {
		t.Fatalf("talosctl not found: %v", err)
	}

	mock := &ProxyMock{}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		app, err := applier.New(ctx, "test-cluster", nil, nil)
		if err != nil {
			return err
		}

		output, err := app.GenerateSecrets(nil)
		assert.NoError(t, err)

		output.ApplyT(func(yaml string) error {
			assert.NotEmpty(t, yaml)
			assert.Contains(t, yaml, "secrets:")
			return nil
		})

		return nil
	}, pulumi.WithMocks("project", "stack", mock))

	assert.NoError(t, err)
	assert.NotEmpty(t, mock.lastStdout)
}

func TestGenerateConfigWithRealTalosctl(t *testing.T) {
	t.Setenv("PULUMI_MOCK_RESOURCES", "1")

	if _, err := exec.LookPath("talosctl"); err != nil {
		t.Fatalf("talosctl not found: %v", err)
	}

	mock := &ProxyMock{}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		app, err := applier.New(ctx, "test-cluster", nil, nil)
		if err != nil {
			return err
		}

		secrets, err := app.GenerateSecrets(nil)
		assert.NoError(t, err)

		cluster := &types.Cluster{
			ClusterName:     "test-cluster",
			ClusterEndpoint: pulumi.String("https://10.0.0.1:6443"),
		}
		machine := &types.ClusterMachine{
			MachineID:     "cp-1",
			MachineType:   "controlplane",
			ConfigPatches: pulumi.StringArray{pulumi.String("")},
		}

		cmd, err := app.GenerateConfig(cluster, machine, secrets)
		assert.NoError(t, err)

		// Capture stdout via the mock for validation.
		cmd.(pulumi.CustomResource).URN().ApplyT(func(_ any) error {
			assert.NotEmpty(t, mock.lastStdout)
			return nil
		})

		return nil
	}, pulumi.WithMocks("project", "stack", mock))

	assert.NoError(t, err)
	assert.NotEmpty(t, mock.lastStdout)
}
