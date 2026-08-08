package provider

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"github.com/stretchr/testify/require"
)

func TestMachineInfoAtOnlyKeepsSensitiveValuesSecret(t *testing.T) {
	machines := pulumi.ToSecret(pulumi.MapArrayMap{
		"init": pulumi.MapArray{
			pulumi.Map{
				types.NodeIPKey:            pulumi.String("192.168.1.12"),
				types.ClusterEnpointKey:    pulumi.String("https://192.168.1.12:6443"),
				types.UserConfigPatchesKey: pulumi.String("machine:\n  token: sensitive"),
				types.TalosImageKey:        pulumi.String("factory.talos.dev/installer:v1.12.9"),
				types.KubernetesVersionKey: pulumi.String("v1.33.5"),
				types.ConfigurationKey:     pulumi.String("machine:\n  token: sensitive"),
			},
		},
	}).(pulumi.MapArrayMapOutput)

	machine := machineInfoAt(machines, "init", 0, "master-1")

	require.False(t, pulumi.IsSecret(machine.NodeIP.ToStringOutput()))
	require.False(t, pulumi.IsSecret(machine.ClusterEnpoint.ToStringOutput()))
	require.False(t, pulumi.IsSecret(machine.TalosImage.ToStringOutput()))
	require.False(t, pulumi.IsSecret(machine.KubernetesVersion.ToStringOutput()))
	require.True(t, pulumi.IsSecret(machine.UserConfigPatches.ToStringOutput()))
	require.True(t, pulumi.IsSecret(machine.Configuration.ToStringOutput()))
}
