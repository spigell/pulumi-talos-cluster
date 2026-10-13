package provider

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
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

func TestMachineTopologyFromApplyMachines(t *testing.T) {
	var (
		initType         = tmachine.TypeInit.String()
		controlPlaneType = tmachine.TypeControlPlane.String()
		workerType       = tmachine.TypeWorker.String()
	)
	machines := pulumi.ToSecret(pulumi.MapArrayMap{
		initType:         pulumi.MapArray{pulumi.Map{types.MachineIDKey: pulumi.String("master-1")}},
		controlPlaneType: pulumi.MapArray{pulumi.Map{types.MachineIDKey: pulumi.String("master-2")}},
		workerType: pulumi.MapArray{
			pulumi.Map{types.MachineIDKey: pulumi.String("worker-1")},
			pulumi.Map{types.MachineIDKey: pulumi.String("worker-2")},
		},
	}).(pulumi.MapArrayMapOutput)

	topology := machineTopologyFromApplyMachines(machines)
	resolved := make(chan map[string][]string, 1)
	topology.ApplyT(func(v map[string][]string) map[string][]string {
		resolved <- v
		return v
	})

	require.False(t, pulumi.IsSecret(topology))
	require.Equal(t, map[string][]string{
		initType:         {"master-1"},
		controlPlaneType: {"master-2"},
		workerType:       {"worker-1", "worker-2"},
	}, <-resolved)
}
