package applier

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"gopkg.in/yaml.v3"
)

// K8SImages holds Kubernetes component images extracted from a Talos configuration.
type K8SImages struct {
	Kubelet           string
	ControllerManager string
	APIServer         string
	KubeProxy         string
	Scheduler         string
}

// MachineConfig represents the parsed YAML structure.
type MachineConfig struct {
	Spec string
}

// NewK8SImages extracts only explicitly configured images from the current machine spec.
func NewK8SImages(spec string) (*K8SImages, error) {
	var current struct {
		Machine struct {
			Type    string `yaml:"type"`
			Kubelet struct {
				Image string `yaml:"image"`
			} `yaml:"kubelet"`
		} `yaml:"machine"`
		Cluster struct {
			APIServer struct {
				Image string `yaml:"image"`
			} `yaml:"apiServer"`
			Proxy struct {
				Image string `yaml:"image"`
			} `yaml:"proxy"`
			Scheduler struct {
				Image string `yaml:"image"`
			} `yaml:"scheduler"`
			ControllerManager struct {
				Image string `yaml:"image"`
			} `yaml:"controllerManager"`
		} `yaml:"cluster"`
	}
	if err := yaml.Unmarshal([]byte(spec), &current); err != nil {
		return nil, fmt.Errorf("error parsing YAML spec string: %w", err)
	}
	images := &K8SImages{Kubelet: current.Machine.Kubelet.Image}
	if current.Machine.Type == machine.TypeControlPlane.String() || current.Machine.Type == machine.TypeInit.String() {
		images.APIServer = current.Cluster.APIServer.Image
		images.KubeProxy = current.Cluster.Proxy.Image
		images.Scheduler = current.Cluster.Scheduler.Image
		images.ControllerManager = current.Cluster.ControllerManager.Image
	}
	return images, nil
}

// apply prepares and returns a Talos CLI command to apply a machine configuration.
// This function merges base machine configuration with user-provided patches and ensures
// that Kubernetes image versions in the configuration align with the currently running
// versions to prevent accidental downgrades (Talos does not support downgrades via specifying images in the config).
func (a *Applier) apply(m *types.MachineInfo, deps []pulumi.Resource) (pulumi.Resource, error) {
	machineFile := pulumi.All(m.UserConfigPatches, m.NodeIP, m.Configuration).ApplyT(func(args []any) (pulumi.StringOutput, error) {
		// Extract current images to use instead of any potential downgraded images
		userPatches := args[0].(string)
		ip := args[1].(string)
		machineConfig := args[2].(string)

		t2 := talosctl.New().
			WithNodeIP(ip).
			WithTalosConfig(a.TalosconfigForNode(ip))
		stageName := "cli-get-machine-config"

		current, err := t2.RunGetCommand(a.ctx, &talosctl.Args{
			Dir:         generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID),
			CommandArgs: pulumi.String("get machineconfig v1alpha1 -oyaml"),
			// No retry. Need to implement another way to retry for get functions.
			RetryCount:  0,
			PrepareDeps: deps,
		})
		if err != nil {
			return pulumi.StringOutput{}, fmt.Errorf("failed to get current machine info: %w", err)
		}

		return current.ApplyT(func(output string) (string, error) {
			var config MachineConfig
			if err := yaml.Unmarshal([]byte(output), &config); err != nil {
				return "", fmt.Errorf("error parsing YAML output: %w", err)
			}

			oldK8SImages, err := NewK8SImages(config.Spec)
			if err != nil {
				return "", err
			}

			// Merge the base machine configuration with user-provided patches.
			// This combines the configs into a single YAML representation.
			merged, err := MergeYAML(machineConfig, userPatches).WithGuard(GuardUnmodifyK8sImages(oldK8SImages)).Build()
			if err != nil {
				return "", fmt.Errorf("failed merge yaml strings: %w", err)
			}

			return merged, nil
		}).(pulumi.StringOutput), nil
	}).(pulumi.StringOutput)

	stageName := "cli-apply-config"

	t := talosctl.New().
		WithNodeIP(m.NodeIP).
		WithTalosConfig(a.TalosconfigForNode(m.NodeIP))

	machineConfigName := "machineconfig.yaml"

	apply, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		AdditionalFiles: []talosctl.ExtraFile{
			{Name: machineConfigName, Content: machineFile},
		},
		CommandArgs: pulumi.Sprintf("apply-config -f %s", machineConfigName),
		Dir:         generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID),
		Triggers: pulumi.Array{
			pulumi.String(m.UserConfigPatches),
			pulumi.String(m.ClusterEnpoint),
		},
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: timeoutShort, Update: timeoutShort}),
		pulumi.DependsOn(deps),
	}...)
	if err != nil {
		return nil, err
	}

	return apply, nil
}
