package applier_test

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"github.com/stretchr/testify/assert"
)

const (
	testClusterName         = "test-cluster"
	controlPlaneMachineType = "controlplane"
	workerMachineType       = "worker"
)

func TestGenerateSecretsWithRealTalosctl(t *testing.T) {
	mock := &ProxyMock{t: t}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		app, err := applier.New(ctx, testClusterName, pulumi.StringMap{}, nil)
		app.WithHooks(false)
		if err != nil {
			return err
		}

		output, err := app.GenerateSecrets()
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

	mock := &ProxyMock{t: t}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		app, err := applier.New(ctx, testClusterName, pulumi.StringMap{}, nil)
		app.WithHooks(false)
		if err != nil {
			return err
		}

		secrets, err := app.GenerateSecrets()
		assert.NoError(t, err)

		cluster := &types.Cluster{
			ClusterName:          testClusterName,
			ClusterEndpoint:      pulumi.String("https://10.0.0.1:6443"),
			KubernetesVersion:    pulumi.String("1.35.0"),
			TalosVersionContract: pulumi.String("v1.12.0"),
		}
		machine := &types.ClusterMachine{
			MachineID:     "cp-1",
			MachineType:   controlPlaneMachineType,
			ConfigPatches: pulumi.StringArray{pulumi.String("")},
			TalosImage:    pulumi.StringPtr("ghcr.io/siderolabs/installer:v1.12.1"),
		}

		cmd, err := app.GenerateMachineConfig(cluster, machine, secrets)
		assert.NoError(t, err)

		cmd.ApplyT(func(raw string) error {
			assert.NotEmpty(t, mock.lastStdout)
			assert.Equal(t, 0, mock.lastExit, "talosctl gen config failed: stdout=%s stderr=%s", mock.lastStdout, mock.lastStderr)
			assert.Contains(t, mock.lastCreate, "--install-image ghcr.io/siderolabs/installer:v1.12.1")
			assert.Contains(t, mock.lastCreate, "--kubernetes-version 1.35.0")
			assert.Contains(t, mock.lastCreate, "--talos-version v1.12.0")
			assert.Contains(t, mock.lastCreate, "--output-types controlplane")
			assert.Contains(t, mock.lastCreate, "--output -")
			assert.Contains(t, raw, "type: controlplane")
			return nil
		})

		return nil
	}, pulumi.WithMocks("project", "stack", mock))

	assert.NoError(t, err)
	assert.NotEmpty(t, mock.lastStdout)
}

func TestGenerateTalosconfigWithRealTalosctl(t *testing.T) {
	t.Setenv("PULUMI_MOCK_RESOURCES", "1")

	mock := &ProxyMock{t: t}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		app, err := applier.New(ctx, testClusterName, pulumi.StringMap{}, nil)
		app.WithHooks(false)
		if err != nil {
			return err
		}

		secrets, err := app.GenerateSecrets()
		assert.NoError(t, err)

		cluster := &types.Cluster{
			ClusterName:          testClusterName,
			ClusterEndpoint:      pulumi.String("https://10.0.0.1:6443"),
			KubernetesVersion:    pulumi.String("1.35.0"),
			TalosVersionContract: pulumi.String("v1.12.0"),
			ClusterMachines: []*types.ClusterMachine{{
				MachineID:     "cp-1",
				MachineType:   controlPlaneMachineType,
				ConfigPatches: pulumi.StringArray{pulumi.String("")},
				TalosImage:    pulumi.StringPtr("ghcr.io/siderolabs/installer:v1.12.1"),
			}},
		}

		output, err := app.GenerateTalosconfig(cluster, secrets)
		assert.NoError(t, err)

		output.ApplyT(func(raw string) error {
			assert.NotEmpty(t, raw)

			cfg, err := clientconfig.FromString(raw)
			assert.NoError(t, err)
			assert.NotEmpty(t, cfg.Contexts)

			ctx := cfg.Contexts[cfg.Context]
			if ctx == nil {
				for _, candidate := range cfg.Contexts {
					ctx = candidate
					break
				}
			}
			assert.NotNil(t, ctx)
			assert.NotEmpty(t, ctx.CA)
			assert.NotEmpty(t, ctx.Crt)
			assert.NotEmpty(t, ctx.Key)

			return nil
		})

		return nil
	}, pulumi.WithMocks("project", "stack", mock))

	assert.NoError(t, err)
	assert.NotEmpty(t, mock.lastStdout)
}

func TestGenerateConfigMachineTypes(t *testing.T) {
	t.Setenv("PULUMI_MOCK_RESOURCES", "1")

	tests := []struct {
		name         string
		machineType  string
		expectTypeIn string
		patches      []string
		expectSnips  []string
	}{
		{name: controlPlaneMachineType, machineType: controlPlaneMachineType, expectTypeIn: "type: controlplane"},
		{name: workerMachineType, machineType: workerMachineType, expectTypeIn: "type: worker"},
		{name: "init", machineType: "init", expectTypeIn: "type: init"},
		{
			name:         "controlplane-with-patch",
			machineType:  controlPlaneMachineType,
			expectTypeIn: "type: controlplane",
			patches: []string{
				// machineBase from hcloud-go cluster.yaml
				"machine:\n  kubelet:\n    nodeIP:\n      validSubnets:\n        - 10.10.10.0/24\n  time:\n    disabled: true\n",
				// controlplaneCluster
				"cluster:\n  etcd:\n    advertisedSubnets:\n      - 10.10.10.0/24\n",
				// timeEnabled overrides
				"machine:\n  time:\n    disabled: false\n",
				// cloudflared extension
				"apiVersion: v1alpha1\nkind: ExtensionServiceConfig\nname: cloudflared\nenvironment:\n  - TUNNEL_TOKEN=CHANGE_ME_AGAIN\n  - TUNNEL_METRICS=localhost:2001\n  - TUNNEL_EDGE_IP_VERSION=auto\n",
			},
			expectSnips: []string{
				"validSubnets:",
				"10.10.10.0/24",
				"advertisedSubnets:",
				"10.10.10.0/24",
				"disabled: false", // from timeEnabled patch
				"kind: ExtensionServiceConfig",
				"name: cloudflared",
			},
		},
		{
			name:         "worker-with-empty-patch",
			machineType:  workerMachineType,
			expectTypeIn: "type: worker",
			patches:      []string{"", "machine:\n  type: worker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &ProxyMock{t: t}
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				app, err := applier.New(ctx, testClusterName, pulumi.StringMap{}, nil)
				app.WithHooks(false)
				if err != nil {
					return err
				}

				secrets, err := app.GenerateSecrets()
				assert.NoError(t, err)

				cluster := &types.Cluster{
					ClusterName:          testClusterName,
					ClusterEndpoint:      pulumi.String("https://10.0.0.1:6443"),
					KubernetesVersion:    pulumi.String("1.35.0"),
					TalosVersionContract: pulumi.String("v1.12.0"),
				}
				machine := &types.ClusterMachine{
					MachineID:     tt.name,
					MachineType:   tt.machineType,
					ConfigPatches: pulumi.ToStringArray(tt.patches),
					TalosImage:    pulumi.StringPtr("ghcr.io/siderolabs/installer:v1.12.1"),
				}

				cmd, err := app.GenerateMachineConfig(cluster, machine, secrets)
				assert.NoError(t, err)

				cmd.ApplyT(func(raw string) error {
					assert.NotEmpty(t, mock.lastStdout)
					assert.Equal(t, 0, mock.lastExit, "talosctl gen config failed: stdout=%s stderr=%s", mock.lastStdout, mock.lastStderr)
					assert.Contains(t, raw, tt.expectTypeIn)
					for _, snip := range tt.expectSnips {
						assert.Contains(t, raw, snip)
					}
					return nil
				})

				return nil
			}, pulumi.WithMocks("project", "stack", mock))

			assert.NoError(t, err)
		})
	}
}
