package applier

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestUpgradeSkipIfNodeRunsTargetImage(t *testing.T) {
	const (
		schematic    = "376567988ad370138ad8b2698212367b8edcb69b5fd68c80be1f2ec7d603b4ba"
		factoryImage = "factory.talos.dev/metal-installer/" + schematic + ":v1.12.9"
	)

	machineConfig := func(image string) string {
		return "node: 10.0.0.2\nmetadata:\n    namespace: config\n    id: v1alpha1\n    version: 3\nspec: |\n" +
			"    version: v1alpha1\n    machine:\n        type: worker\n        install:\n            disk: /dev/sda\n" +
			"            image: " + image + "\n            wipe: false\n        kubelet:\n            image: ghcr.io/siderolabs/kubelet:v1.33.5\n"
	}
	version := func(tag string) string {
		return "Client:\n\tTag:         v1.12.9\nServer:\n\tNODE:        10.0.0.2\n\tTag:         " + tag + "\n"
	}
	extensions := func(id string) string {
		return "node: 10.0.0.2\nmetadata:\n    namespace: runtime\n    id: \"0\"\n    version: 1\nspec:\n    image: 0.sqsh\n" +
			"    metadata:\n        name: schematic\n        version: " + id + "\n        author: Image Factory\n"
	}

	tests := []struct {
		name          string
		target        string
		machineConfig string
		version       string
		extensions    string
		skip          bool
	}{
		{
			name: "factory image already running", target: factoryImage,
			machineConfig: machineConfig(factoryImage), version: version("v1.12.9"), extensions: extensions(schematic), skip: true,
		},
		{
			name: "plain installer already running", target: "ghcr.io/siderolabs/installer:v1.12.9",
			machineConfig: machineConfig("ghcr.io/siderolabs/installer:v1.12.9"), version: version("v1.12.9"), skip: true,
		},
		{
			name: "running tag differs", target: factoryImage,
			machineConfig: machineConfig(factoryImage), version: version("v1.12.8"), extensions: extensions(schematic),
		},
		{
			name: "running schematic differs", target: factoryImage,
			machineConfig: machineConfig(factoryImage), version: version("v1.12.9"), extensions: extensions(strings.Repeat("0", 64)),
		},
		{
			name: "installed config has a stale image", target: factoryImage,
			machineConfig: machineConfig("ghcr.io/siderolabs/installer:v1.12.9"), version: version("v1.12.9"), extensions: extensions(schematic),
		},
		{
			name: "image without tag", target: "registry.local:5000/installer",
			machineConfig: machineConfig("registry.local:5000/installer"), version: version("v1.12.9"),
		},
		{
			name: "node unreachable", target: factoryImage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range map[string]string{"machineconfig": tt.machineConfig, "version": tt.version, "extensions": tt.extensions} {
				require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
			}
			fake := "#!/bin/bash\ncase \"$*\" in\n" +
				"  *'get machineconfig'*) f=machineconfig ;;\n  *'get extensions'*) f=extensions ;;\n  *version*) f=version ;;\nesac\n" +
				"[ -s \"$FAKE_DIR/$f\" ] || exit 1\ncat \"$FAKE_DIR/$f\"\n"
			require.NoError(t, os.WriteFile(filepath.Join(dir, "talosctl"), []byte(fake), 0o700))

			cmd := exec.CommandContext(t.Context(), "/bin/bash", "-c", "talosctl_node() { talosctl -n 10.0.0.2 \"$@\"; }\n"+upgradeSkipIfNodeRunsTargetImage)
			cmd.Env = append(os.Environ(),
				"PATH="+dir+":"+os.Getenv("PATH"),
				"FAKE_DIR="+dir,
				"NODE_IP=10.0.0.2",
				"TALOS_TARGET_IMAGE="+tt.target,
			)
			out, err := cmd.CombinedOutput()

			if tt.skip {
				require.NoError(t, err, string(out))
				require.Contains(t, string(out), "skipping talosctl upgrade")
			} else {
				require.Error(t, err, string(out))
			}
		})
	}
}
