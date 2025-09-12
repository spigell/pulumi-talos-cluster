package applier

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"gopkg.in/yaml.v3"
)

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

func NewK8SImages(config *v1alpha1.Config) *K8SImages {
	images := &K8SImages{
		Kubelet: config.MachineConfig.MachineKubelet.KubeletImage,
	}

	// This struct is not filled in worker configurations.
	if config.MachineConfig.MachineType == machine.TypeControlPlane.String() || config.MachineConfig.MachineType == machine.TypeInit.String() {
		images.APIServer = config.ClusterConfig.APIServerConfig.ContainerImage
		images.KubeProxy = config.ClusterConfig.ProxyConfig.ContainerImage
		images.Scheduler = config.ClusterConfig.SchedulerConfig.ContainerImage
		images.ControllerManager = config.ClusterConfig.ControllerManagerConfig.ContainerImage
	}

	return images
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

		t2 := talosctl.New().WithNodeIP(ip)
		stageName := "cli-get-machine-config"

		current, err := t2.RunGetCommand(a.ctx, stageName, &talosctl.TalosctlArgs{
			TalosConfig: a.basicClient().TalosConfig(),
			Dir: generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID),
			CommandArgs: pulumi.String("get machineconfig v1alpha1 -oyaml"),
			// No retry. Need to implement another way to retry for get functions.
			RetryCount: 0,
		}, deps)

		if err != nil {
			return pulumi.StringOutput{}, fmt.Errorf("failed to get current machine info: %w", err)
		}


		return current.ApplyT(func (output string) (string, error) {
			var config MachineConfig
			if err := yaml.Unmarshal([]byte(output), &config); err != nil {
				return "", fmt.Errorf("error parsing YAML output: %w", err)
			}

			var spec v1alpha1.Config
			if err := yaml.Unmarshal([]byte(config.Spec), &spec); err != nil {
				return "", fmt.Errorf("error parsing YAML spec string: %w", err)
			}

			oldK8SImages := NewK8SImages(&spec)

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
	t := talosctl.New().WithNodeIP(m.NodeIP)
	machineConfigName := "machineconfig.yaml"

	apply, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.TalosctlArgs{
		TalosConfig: a.basicClient().TalosConfig(),
		AdditionalFiles: []talosctl.ExtraFile{
			{ Name: machineConfigName, Content: machineFile },
		},
		CommandArgs: pulumi.Sprintf("apply-config -f %s", machineConfigName),
		Dir: generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID),
		Triggers:    pulumi.Array{
			pulumi.String(m.UserConfigPatches),
			pulumi.String(m.ClusterEnpoint),
		},
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "90s", Update: "90s"}),
		pulumi.DependsOn(deps),
	}...)

	if err != nil {
		return nil, err
	}

	return apply, nil
}
