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

// NewK8SImages constructs a K8SImages structure from the provided Talos config.
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
	machineFile := pulumi.All(m.NodeIP, m.Configuration, m.TalosImage).ApplyT(func(args []any) (pulumi.StringOutput, error) {
		// Extract current images to use instead of any potential downgraded images
		ip := args[0].(string)
		machineConfig := args[1].(string)
		talosImage := args[2].(string)

		t2 := talosctl.New().
			WithNodeIP(ip).
			WithTalosConfig(a.TalosconfigForNode(pulumi.String(ip)))
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

			var spec v1alpha1.Config
			if err := yaml.Unmarshal([]byte(config.Spec), &spec); err != nil {
				return "", fmt.Errorf("error parsing YAML spec string: %w", err)
			}

			oldK8SImages := NewK8SImages(&spec)

			merged, err := buildSecureApplyConfig(machineConfig, talosImage, oldK8SImages)
			if err != nil {
				return "", fmt.Errorf("failed merge yaml strings: %w", err)
			}

			return merged, nil
		}).(pulumi.StringOutput), nil
	}).(pulumi.StringOutput)

	stageName := "cli-apply-config"

	t := talosctl.New().
		WithNodeIPInput(m.NodeIP).
		WithTalosConfig(a.TalosconfigForNode(m.NodeIP))

	machineConfigName := "machineconfig.yaml"

	apply, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		AdditionalFiles: []talosctl.ExtraFile{
			{Name: machineConfigName, Content: machineFile},
		},
		CommandArgs:    pulumi.Sprintf("apply-config -f %s", machineConfigName),
		Dir:            generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID),
		UpdateOnChange: true,
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

func buildSecureApplyConfig(machineConfig, talosImage string, oldK8SImages *K8SImages) (string, error) {
	installImagePatch := fmt.Sprintf("machine:\n  install:\n    image: %s\n", talosImage)
	return MergeYAML(machineConfig, installImagePatch).WithGuard(GuardUnmodifyK8sImages(oldK8SImages)).Build()
}
