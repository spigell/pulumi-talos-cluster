package applier

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
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

// talosctlApplyCMD prepares and returns a Talos CLI command to apply a machine configuration.
// This function merges base machine configuration with user-provided patches and ensures
// that Kubernetes image versions in the configuration align with the currently running
// versions to prevent accidental downgrades (Talos does not support downgrades via specifying images in the config).
func (a *Applier) talosctlApplyCMD(m *types.MachineInfo) pulumi.StringOutput {
	return pulumi.All(a.basicClient().TalosConfig(), m.UserConfigPatches, m.NodeIP, m.Configuration).ApplyT(func(args []any) (string, error) {
		// Unpack asynchronous values required for the configuration.
		talosConfig := args[0].(string)
		userPatches := args[1].(string)
		ip := args[2].(string)
		machineConfig := args[3].(string)

		// Initialize the Talos CLI and prepare a temporary home directory.
		var config v1alpha1.Config
		talosctl := a.NewTalosctl()
		if err := talosctl.prepare(talosConfig); err != nil {
			return "", fmt.Errorf("failed to prepare temp home for talos cli: %w", err)
		}

		// Merge the base machine configuration with user-provided patches.
		// This combines the configs into a single YAML representation.
		merged, err := mergeYAML(machineConfig, userPatches)
		if err != nil {
			return "", fmt.Errorf("failed merge yaml strings: %w", err)
		}

		err = yaml.Unmarshal([]byte(merged), &config)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal config from string: %w", err)
		}

		// Initialize new Kubernetes image set from the current configuration.
		// In the dry-run mode this values will be used in initial configuration files.
		newK8SImages := NewK8SImages(&config)

		// If not in dry-run mode, retrieve the current machine configuration.
		// This function will run in async way with retry so we hope that the init node will be ready soon.
		// Not very safe way.
		if !a.ctx.DryRun() {
			current, err := talosctl.getCurrentMachineConfig(a.InitNode.IP)
			if err != nil {
				return "", fmt.Errorf("failed to get current machine info: %w", err)
			}

			// Extract current images to use instead of any potential downgraded images
			oldK8SImages := NewK8SImages(current)
			a.ctx.Log.Debug(fmt.Sprintf("overwriting k8s images version %+v with %+v", newK8SImages, oldK8SImages), nil)
			newK8SImages = oldK8SImages
		}

		config.MachineConfig.MachineKubelet.KubeletImage = newK8SImages.Kubelet
		if config.MachineConfig.MachineType == machine.TypeControlPlane.String() || config.MachineConfig.MachineType == machine.TypeInit.String() {
			config.ClusterConfig.APIServerConfig.ContainerImage = newK8SImages.APIServer
			config.ClusterConfig.ProxyConfig.ContainerImage = newK8SImages.KubeProxy
			config.ClusterConfig.ControllerManagerConfig.ContainerImage = newK8SImages.ControllerManager
			config.ClusterConfig.SchedulerConfig.ContainerImage = newK8SImages.Scheduler
		}

		// Marshal the modified configuration back to YAML format to write to a file
		marshalled, err := yaml.Marshal(&config)
		if err != nil {
			return "", fmt.Errorf("failed to marshal merged YAML: %w", err)
		}

		// Write the marshalled YAML to a temporary file for Talos CLI to apply
		configPath := filepath.Join(talosctl.Home.Dir, fmt.Sprintf("machineconfig-%s.yaml", m.MachineID))
		if err := os.WriteFile(configPath, marshalled, 0o600); err != nil {
			return "", fmt.Errorf("failed to write machine config: %w", err)
		}

		// Construct the Talos CLI command to apply the configuration to the machine
		// `withBashRetry` ensures command retry in case the machine isn't ready yet
		command := withBashRetry(fmt.Sprintf(strings.Join([]string{
			"%[1]s apply-config -n %[2]s -e %[2]s -f %s",
		}, " && "), talosctl.BasicCommand, ip, configPath), "5")

		return command, nil
	}).(pulumi.StringOutput)
}
