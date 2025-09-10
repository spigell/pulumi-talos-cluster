package applier

import (
	"fmt"
	"runtime"

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
	stageName := "apply-config"
	home := generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID)
	t := talosctl.New(a.ctx, home, deps)

	machineFile := pulumi.All(m.UserConfigPatches, m.NodeIP, m.Configuration).ApplyT(func(args []any) (pulumi.StringOutput, error) {
		runtime.Breakpoint()
		// Extract current images to use instead of any potential downgraded images
		userPatches := args[0].(string)
		ip := args[1].(string)
		machineConfig := args[2].(string)



		current, err := t.RunGetCommand("get-machine-config", &talosctl.TalosctlArgs{
			TalosConfig: a.basicClient().TalosConfig(),
			Args: pulumi.String(fmt.Sprintf("%s get machineconfig v1alpha1 -n %[2]s -e %[2]s -oyaml",
				t.BasicCommand, ip,
			)),
			RetryCount: 10,
		}, deps)
		//current, err := talosctl.getCurrentMachineConfig(a.InitNode.IP, deps)

		if err != nil {
			return pulumi.StringOutput{}, fmt.Errorf("failed to get current machine info: %w", err)
		}


		return current.ApplyT(func (output string) (string, error) {
			runtime.Breakpoint()
			
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

	configPath := "machineconfig.yaml"

	apply, err := t.RunCommand(fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.TalosctlArgs{
		TalosConfig: a.basicClient().TalosConfig(),
		AdditionalFiles: []talosctl.ExtraFile{
			{ Path: configPath, Content: machineFile },
		},
		Args: pulumi.Sprintf("apply-config -n %[1]s -e %[1]s -f %s", m.NodeIP, configPath),
		RetryCount: 10,
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
