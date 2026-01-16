package provider

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/siderolabs/talos/pkg/machinery/gendata"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

const (
	// DefaultK8SVersion.
	DefaultK8SVersion = "v1.33.0"
)

const (
	ClusterResourceOutputsMachines                                = "machines"
	ClusterResourceOutputsGeneratedConfigurations                 = "generatedConfigurations"
	ClusterResourceOutputsControlplaneMachineConfigurations       = "controlplaneMachineConfigurations"
	ClusterResourceOutputsWorkerMachineConfigurations             = "workerMachineConfigurations"
	ClusterResourceOutputsInitMachineConfiguration                = "initMachineConfiguration"
	ClusterResourceOutputsClientConfiguration                     = "clientConfiguration"
	ClusterResourceOutputsClientConfigurationCAKey                = "caCertificate"
	ClusterResourceOutputsClientConfigurationClientKey            = "clientKey"
	ClusterResourceOutputsClientConfigurationClientCertificateKey = "clientCertificate"
)

type Cluster struct {
	pulumi.ResourceState
	ClusterArgs

	ClientConfiguration     pulumi.StringMap `pulumi:"clientConfiguration"`
	GeneratedConfigurations pulumi.StringMap `pulumi:"generatedConfigurations"`
	Machines                pulumi.ArrayMap  `pulumi:"machines"`
}

func ClusterType() string {
	return ProviderName + ":index:Cluster"
}

type ClusterArgs struct {
	ClusterName          string             `pulumi:"clusterName"`
	TalosVersionContract pulumi.StringInput `pulumi:"talosVersionContract"`
	ClusterEndpoint      pulumi.StringInput `pulumi:"clusterEndpoint"`
	KubernetesVersion    pulumi.StringInput `pulumi:"kubernetesVersion"`

	ClusterMachines []*types.ClusterMachine `pulumi:"clusterMachines"`
}

func GenerateDefaultInstallerImage() string {
	return fmt.Sprintf("%s/%s/installer:%s", gendata.ImagesRegistry, gendata.ImagesUsername, gendata.VersionTag)
}

func cluster(ctx *pulumi.Context, c *Cluster, name string,
	args *ClusterArgs, inputs provider.ConstructInputs, opts ...pulumi.ResourceOption,
) (*provider.ConstructResult, error) {
	// Blit the inputs onto the arguments struct.
	if err := inputs.CopyTo(args); err != nil {
		return nil, errors.Wrap(err, "setting args")
	}

	// Register our component resource.
	if err := ctx.RegisterComponentResource(ClusterType(), name, c, opts...); err != nil {
		return nil, err
	}
	app, err := applier.New(ctx, name,
		nil,
		pulumi.Parent(c),
	)

	// Generate secrets via talosctl.
	secrets, err := app.GenerateSecrets(nil)
	if err != nil {
		return nil, errors.Wrap(err, "generating secrets")
	}

	secretsStash, err := pulumi.NewStash(ctx, fmt.Sprintf("%s:secrets", name), &pulumi.StashArgs{
		Input: secrets,
	}, pulumi.Parent(c))
	if err != nil {
		return nil, errors.Wrap(err, "stashing secrets")
	}

	// Generate configs via talosctl using the generated secrets.
	configCmd, err := applier.GenerateConfig(ctx, name, workDir, args.ClusterName, args.ClusterEndpoint, []pulumi.Resource{secretsCmd}, pulumi.Parent(c))
	if err != nil {
		return nil, errors.Wrap(err, "generating configs")
	}

	workers := make(pulumi.Array, 0)
	controlplanes := make(pulumi.Array, 0)
	generated := make(pulumi.StringMap, 0)
	c.Machines = make(pulumi.ArrayMap)

	for _, m := range args.ClusterMachines {
		// The provider doesn't know anything about init node type.
		if m.ConfigPatches == nil {
			m.ConfigPatches = pulumi.StringArray{pulumi.String("")}
		}

		// Required and Defaults do not work for nested structs in Components?
		if m.TalosImage == nil {
			m.TalosImage = pulumi.String(GenerateDefaultInstallerImage())
		}

		machineType := m.MachineType
		if m.MachineType == tmachine.TypeInit.String() {
			// Use controlplane config for init as before.
			machineType = tmachine.TypeControlPlane.String()
		}

		var cfgFile string
		switch machineType {
		case tmachine.TypeControlPlane.String():
			cfgFile = "controlplane.yaml"
		case tmachine.TypeWorker.String():
			cfgFile = "worker.yaml"
		default:
			return nil, fmt.Errorf("unknown machine type %s", m.MachineType)
		}

		cfg := talosctl.New().CatFile(ctx, filepath.Join(workDir, "configs"), cfgFile, []pulumi.Resource{configCmd})

		generated[m.MachineID] = cfg

		mInfo := m.ToMachineInfoMap(args.ClusterEndpoint, args.KubernetesVersion, cfg)

		switch machineType {
		case tmachine.TypeControlPlane.String():
			controlplanes = append(controlplanes, mInfo)
		case tmachine.TypeWorker.String():
			workers = append(workers, mInfo)
		case tmachine.TypeInit.String():
			if len(c.Machines) == 1 {
				return nil, fmt.Errorf("only one init node should present. Please use 'controlplane' type for %s", m.MachineID)
			}

			c.Machines[tmachine.TypeInit.String()] = pulumi.Array{mInfo}
		default:
			return nil, fmt.Errorf("unknown machine type %s", m.MachineType)
		}
	}

	c.Machines[tmachine.TypeWorker.String()] = workers
	c.Machines[tmachine.TypeControlPlane.String()] = controlplanes

	c.GeneratedConfigurations = generated

	talosconfig := talosctl.New().CatFile(ctx, filepath.Join(workDir, "configs"), "talosconfig", []pulumi.Resource{configCmd})

	c.ClientConfiguration = pulumi.StringMap{
		ClusterResourceOutputsClientConfigurationCAKey:                pulumi.String(""),
		ClusterResourceOutputsClientConfigurationClientKey:            pulumi.String(""),
		ClusterResourceOutputsClientConfigurationClientCertificateKey: pulumi.String(""),
		"talosconfig": talosconfig,
		"secrets":     secretsContent,
	}

	if err := ctx.RegisterResourceOutputs(c, pulumi.Map{
		ClusterResourceOutputsClientConfiguration:     c.ClientConfiguration,
		ClusterResourceOutputsMachines:                c.Machines,
		ClusterResourceOutputsGeneratedConfigurations: generated,
		"secretsStash": secretsStash.Output,
	}); err != nil {
		return nil, err
	}

	return provider.NewConstructResult(c)
}
