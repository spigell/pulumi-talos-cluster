package applier

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTalosctlUpgradeArgsUsesMachineTalosImage(t *testing.T) {
	args, err := talosctlUpgradeArgs("factory.talos.dev/metal-installer/new-image:v1.12.9", "worker-1")

	require.NoError(t, err)
	require.Equal(t, "upgrade --debug --drain=false --image factory.talos.dev/metal-installer/new-image:v1.12.9", args)
}

func TestTalosctlUpgradeArgsRequiresTalosImage(t *testing.T) {
	_, err := talosctlUpgradeArgs("", "worker-1")

	require.Error(t, err)
	require.Contains(t, err.Error(), "talos image is required for machine worker-1")
}
