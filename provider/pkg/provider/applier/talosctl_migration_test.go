package applier

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"github.com/stretchr/testify/require"
)

type commandMocks struct {
	mu     sync.Mutex
	inputs map[string]resource.PropertyMap
}

func (m *commandMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inputs[args.Name] = args.Inputs

	return args.Name + "-id", args.Inputs, nil
}

func (m *commandMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	stdout := ""
	if command := plainString(args.Args["command"]); strings.Contains(command, "get machineconfig") {
		stdout = "spec: |\n  machine:\n    type: worker\n    kubelet:\n      image: ghcr.io/siderolabs/kubelet:v1.33.5\n"
	}

	return resource.PropertyMap{
		"stdout": resource.NewStringProperty(stdout),
		"stderr": resource.NewStringProperty(""),
	}, nil
}

func plainString(v resource.PropertyValue) string {
	if v.IsSecret() {
		v = v.SecretValue().Element
	}
	if !v.IsString() {
		return ""
	}

	return v.StringValue()
}

func plainStrings(v resource.PropertyValue) []string {
	if v.IsSecret() {
		v = v.SecretValue().Element
	}
	if !v.IsArray() {
		return nil
	}

	values := make([]string, 0, len(v.ArrayValue()))
	for _, item := range v.ArrayValue() {
		values = append(values, plainString(item))
	}

	return values
}

// Stacks created with 0.4.1 store triggers for these commands. pulumi-command replaces a command on any
// trigger change and a replace runs create, so the first `pulumi up` after the bump must keep the same
// triggers, and create must be guarded the same way as update in case a replace happens anyway.
func TestCommandsKeep041TriggersAndGuardCreate(t *testing.T) {
	const (
		image      = "ghcr.io/siderolabs/installer:v1.12.9"
		patches    = "machine:\n  network:\n    hostname: worker-1\n"
		endpoint   = "https://10.0.0.1:6443"
		k8sVersion = "1.33.5"
	)

	mocks := &commandMocks{inputs: map[string]resource.PropertyMap{}}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		a, err := New(ctx, "cluster", pulumi.StringMap{
			"caCertificate":     pulumi.String("ca"),
			"clientKey":         pulumi.String("key"),
			"clientCertificate": pulumi.String("cert"),
		}, pulumi.Protect(false))
		if err != nil {
			return err
		}
		a.WithHooks(false)

		m := &types.MachineInfo{
			MachineID:         "worker-1",
			NodeIP:            pulumi.String("10.0.0.2"),
			ClusterEnpoint:    pulumi.String(endpoint),
			UserConfigPatches: pulumi.String(patches),
			TalosImage:        pulumi.String(image),
			KubernetesVersion: pulumi.String(k8sVersion),
			Configuration:     pulumi.String("machine:\n  type: worker\n  install:\n    image: " + image + "\n"),
		}
		if _, err := a.upgrade(m, tmachine.TypeWorker, nil); err != nil {
			return err
		}
		if _, err := a.apply(m, nil); err != nil {
			return err
		}
		_, err = a.upgradeK8S(m, nil)

		return err
	}, pulumi.WithMocks("project", "stack", mocks))
	require.NoError(t, err)

	upgrade := mocks.inputs["cluster:cli-upgrade:worker-1"]
	applyConfig := mocks.inputs["cluster:cli-apply-config:worker-1"]
	upgradeK8S := mocks.inputs["cluster:cli-upgrade-k8s:worker-1"]
	require.NotNil(t, upgrade)
	require.NotNil(t, applyConfig)
	require.NotNil(t, upgradeK8S)

	require.Equal(t, []string{image}, plainStrings(upgrade["triggers"]))
	require.Equal(t, []string{patches, endpoint}, plainStrings(applyConfig["triggers"]))
	require.Equal(t, []string{k8sVersion}, plainStrings(upgradeK8S["triggers"]))

	for name, inputs := range map[string]resource.PropertyMap{"cli-upgrade": upgrade, "cli-apply-config": applyConfig} {
		create := plainString(inputs["create"])
		require.Contains(t, create, "talosctl_node()", "%s create must run the skip guard", name)
		require.Equal(t, create, plainString(inputs["update"]), "%s create and update must match", name)
	}
	require.Contains(t, plainString(applyConfig["create"]), "apply-config --dry-run -f machineconfig.yaml")

	// Run the registered create command, as a replace would, against a node that already runs the image
	// and against one that does not.
	machineConfig := "spec: |\n    machine:\n        install:\n            image: " + image + "\n"
	for _, tc := range []struct {
		name        string
		runningTag  string
		wantUpgrade bool
	}{
		{name: "node already runs the image", runningTag: "v1.12.9"},
		{name: "node runs an older image", runningTag: "v1.12.8", wantUpgrade: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeFakeTalosctl(t, machineConfig, "Server:\n\tTag:         "+tc.runningTag+"\n", "")

			cmd := exec.CommandContext(t.Context(), "/bin/bash", "-c", plainString(upgrade["create"]))
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "FAKE_DIR="+dir)
			for key, value := range upgrade["environment"].ObjectValue() {
				cmd.Env = append(cmd.Env, string(key)+"="+plainString(value))
			}
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, string(out))

			calls, err := os.ReadFile(dir + "/calls")
			require.NoError(t, err)
			require.Equal(t, tc.wantUpgrade, strings.Contains(string(calls), "upgrade --debug --drain=false --image "+image), string(calls))
		})
	}
}
