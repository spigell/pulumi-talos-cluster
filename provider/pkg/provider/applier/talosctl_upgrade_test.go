package applier

import (
	"testing"

	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"github.com/stretchr/testify/require"
)

func TestTalosctlUpgradeArgsUsesMachineTalosImage(t *testing.T) {
	machine := &types.MachineInfo{
		MachineID:     "worker-1",
		TalosImage:    "factory.talos.dev/metal-installer/new-image:v1.12.9",
		Configuration: "machine:\n  install:\n    image: factory.talos.dev/metal-installer/old-image:v1.12.1\n",
	}

	args, err := talosctlUpgradeArgs(machine)

	require.NoError(t, err)
	require.Equal(t, "upgrade --debug --drain=false --image factory.talos.dev/metal-installer/new-image:v1.12.9", args)
}

func TestTalosctlUpgradeArgsRequiresTalosImage(t *testing.T) {
	machine := &types.MachineInfo{MachineID: "worker-1"}

	_, err := talosctlUpgradeArgs(machine)

	require.Error(t, err)
	require.Contains(t, err.Error(), "talos image is required for machine worker-1")
}
