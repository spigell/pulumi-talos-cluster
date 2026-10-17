package applier

import (
	"testing"

	"github.com/siderolabs/talos/pkg/machinery/config"
	"github.com/siderolabs/talos/pkg/machinery/config/configpatcher"
	"github.com/siderolabs/talos/pkg/machinery/config/generate"
	"github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/stretchr/testify/require"
)

func generatedConfig(t *testing.T, patches ...string) string {
	t.Helper()

	contract, err := config.ParseContractFromVersion("v1.14.2")
	require.NoError(t, err)

	// Like `talosctl gen config`, whose default install disk makes it add the UnattendedInstallConfig document.
	in, err := generate.NewInput("test", "https://10.0.0.1:6443", "1.32.0",
		generate.WithVersionContract(contract),
		generate.WithInstallDisk("/dev/sda"),
		generate.WithInstallImage("ghcr.io/siderolabs/installer:v1.14.2"),
	)
	require.NoError(t, err)

	cfg, err := in.Config(machine.TypeWorker)
	require.NoError(t, err)

	loaded, err := configpatcher.LoadPatches(patches)
	require.NoError(t, err)

	out, err := configpatcher.Apply(configpatcher.WithConfig(cfg), loaded)
	require.NoError(t, err)

	raw, err := out.Bytes()
	require.NoError(t, err)

	return string(raw)
}

func TestValidateMachineConfig(t *testing.T) {
	t.Run("generated config is valid", func(t *testing.T) {
		require.NoError(t, validateMachineConfig("worker-1", generatedConfig(t)))
	})

	t.Run("KubeNodeConfig patch is valid", func(t *testing.T) {
		raw := generatedConfig(t, "apiVersion: v1alpha1\nkind: KubeNodeConfig\nnodeIP:\n  validSubnets:\n    - 10.10.10.0/24\n")
		require.NoError(t, validateMachineConfig("worker-1", raw))
	})

	t.Run("legacy machine.kubelet patch conflicts with the v1.14 contract", func(t *testing.T) {
		raw := generatedConfig(t, "machine:\n  kubelet:\n    nodeIP:\n      validSubnets:\n        - 10.10.10.0/24\n")
		err := validateMachineConfig("worker-1", raw)
		require.ErrorContains(t, err, "machine worker-1: generated machine config is invalid, check its configPatches")
		require.ErrorContains(t, err, "kubelet config is already set in v1alpha1 config")
	})

	t.Run("garbage cannot be loaded", func(t *testing.T) {
		require.ErrorContains(t, validateMachineConfig("worker-1", "not: [a talos config"), "cannot be loaded")
	})

	t.Run("empty output is skipped", func(t *testing.T) {
		require.NoError(t, validateMachineConfig("worker-1", ""))
	})
}

func TestBuildSecureApplyConfigWithV114Contract(t *testing.T) {
	const image = "factory.talos.dev/installer/abc:v1.14.2"

	out, err := buildSecureApplyConfig(generatedConfig(t), image, &K8SImages{})
	require.NoError(t, err)
	require.NoError(t, validateMachineConfig("worker-1", out), "the node would reject this config")

	require.Contains(t, out, "kind: UnattendedInstallConfig")
	require.Contains(t, out, "image: "+image)
	require.NotContains(t, out, "\n    install:", "machine.install conflicts with UnattendedInstallConfig")
}
