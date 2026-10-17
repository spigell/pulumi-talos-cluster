package applier

import (
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

// The only deployed stack was created from the PR #39 branch build (f946658 + 38c95ab), which stores no
// triggers and plain create/update commands. pulumi-command replaces a command on any trigger change and a
// replace re-runs create, so these commands must not send triggers, and create must stay the plain command
// so that state matches and nothing re-runs unless its inputs really change.
func TestCommandsUpdateInPlaceWithoutTriggers(t *testing.T) {
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

	for name, inputs := range map[string]resource.PropertyMap{
		"cli-upgrade":      upgrade,
		"cli-apply-config": applyConfig,
		"cli-upgrade-k8s":  upgradeK8S,
	} {
		triggers, ok := inputs["triggers"]
		require.True(t, !ok || triggers.IsNull() || (triggers.IsArray() && len(triggers.ArrayValue()) == 0),
			"%s must not send triggers, got %v", name, triggers)

		create := plainString(inputs["create"])
		require.NotEmpty(t, create, name)
		require.Equal(t, create, plainString(inputs["update"]), "%s create and update must match", name)
		require.NotContains(t, create, "talosctl_node", "%s must run the plain talosctl command", name)
	}

	require.True(t, strings.HasPrefix(plainString(upgrade["create"]),
		"n=0 ; until [ $n -ge 11 ] ; do talosctl -n 10.0.0.2 -e 10.0.0.2 upgrade --debug --drain=false --image "+image+" "),
		plainString(upgrade["create"]))
	require.NotContains(t, upgrade["environment"].ObjectValue(), resource.PropertyKey("TALOS_TARGET_IMAGE"))

	require.Equal(t, "talosctl -n 10.0.0.2 -e 10.0.0.2 apply-config -f machineconfig.yaml --talosconfig talosctl.yaml",
		plainString(applyConfig["create"]))
	require.Contains(t, applyConfig["environment"].ObjectValue(), resource.PropertyKey("PULUMI_TALOS_ADDITIONAL_FILES_SHA256"))
}
