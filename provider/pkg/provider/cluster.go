package provider

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/siderolabs/talos/pkg/machinery/gendata"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultK8SVersion.
	DefaultK8SVersion = "v1.31.0"
)

var (
	ClusterResourceOutputsControlplaneMachineConfigurations       = "controlplaneMachineConfigurations"
	ClusterResourceOutputsWorkerMachineConfigurations             = "workerMachineConfigurations"
	ClusterResourceOutputsInitMachineConfiguration                = "initMachineConfiguration"
	ClusterResourceOutputsUserConfigPatches                       = "userConfigPatches"
	ClusterResourceOutputsClientConfiguration                     = "clientConfiguration"
	ClusterResourceOutputsClientConfigurationCAKey                = "caCertificate"
	ClusterResourceOutputsClientConfigurationClientKey            = "clientKey"
	ClusterResourceOutputsClientConfigurationClientCertificateKey = "clientCertificate"
)

type Cluster struct {
	pulumi.ResourceState
	ClusterArgs

	ClientConfiguration               pulumi.StringMap    `pulumi:"clientConfiguration"`
	InitMachineConfiguration          pulumi.StringOutput `pulumi:"initMachineConfiguration"`
	ControlplaneMachineConfigurations pulumi.Map          `pulumi:"controlplaneMachineConfigurations"`
	WorkerMachineConfigurations       pulumi.Map          `pulumi:"workerMachineConfigurations"`
	UserConfigPatches                 pulumi.Map          `pulumi:"userConfigPatches"`
}

func ClusterType() string {
	return ProviderName + ":index:Cluster"
}

type ClusterArgs struct {
	ClusterName          string             `pulumi:"clusterName"`
	TalosVersionContract pulumi.StringInput `pulumi:"talosVersionContract"`
	ClusterEndpoint      pulumi.StringInput `pulumi:"clusterEndpoint"`

	ClusterMachines []ClusterMachine `pulumi:"clusterMachines"`
}

type ClusterMachine struct {
	MachineID         string                `pulumi:"machineId"`
	MachineType       string                `pulumi:"machineType"`
	TalosImage        pulumi.StringPtrInput `pulumi:"talosImage"`
	KubernetesVersion pulumi.StringPtrInput `pulumi:"kubernetesVersion"`
	ConfigPatches     pulumi.StringPtrInput `pulumi:"configPatches"`
}

type MachineConfiguration struct {
	Configuration     pulumi.StringOutput `pulumi:"configuration"`
	UserConfigPatches pulumi.StringOutput `pulumi:"userConfigPatches"`
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

	secrets, err := machine.NewSecrets(ctx, fmt.Sprintf("%s:secrets", name), &machine.SecretsArgs{
		TalosVersion: args.TalosVersionContract,
	}, pulumi.Parent(c), pulumi.IgnoreChanges([]string{"talosVersion"}))
	if err != nil {
		return nil, err
	}

	workers := make(pulumi.Map, 0)
	controlplanes := make(pulumi.Map, 0)
	userPatches := make(pulumi.Map, 0)

	for _, m := range args.ClusterMachines {
		// The provider doesn't know anything about init node type.
		// It should be the controlplane for it.
		machineType := m.MachineType
		if m.MachineType == tmachine.TypeInit.String() {
			machineType = tmachine.TypeControlPlane.String()
		}

		if m.ConfigPatches == nil {
			m.ConfigPatches = pulumi.String("{}")
		}

		// Required and Defaults do not work for nested structs in Components?
		if m.TalosImage == nil {
			m.TalosImage = pulumi.String(GenerateDefaultInstallerImage())
		}

		if m.KubernetesVersion == nil {
			m.KubernetesVersion = pulumi.String(DefaultK8SVersion)
		}
		configuration := machine.GetConfigurationOutput(ctx, machine.GetConfigurationOutputArgs{
			ClusterName:       pulumi.String(args.ClusterName),
			MachineType:       pulumi.String(machineType),
			ClusterEndpoint:   args.ClusterEndpoint,
			KubernetesVersion: m.KubernetesVersion,
			TalosVersion:      compareContractVersionWithNotify(ctx, secrets.TalosVersion, args.TalosVersionContract.ToStringOutput()),
			ConfigPatches: pulumi.StringArray{
				m.ConfigPatches.ToStringPtrOutput().Elem(),
				configureTalosInstall(m.TalosImage.ToStringPtrOutput().Elem()),
			},
			MachineSecrets: secrets.ToSecretsOutput().MachineSecrets(),
		}, nil)

		userPatches[m.MachineID] = m.ConfigPatches.ToStringPtrOutput().Elem()

		switch m.MachineType {
		case tmachine.TypeControlPlane.String():
			controlplanes[m.MachineID] = configuration.MachineConfiguration()
		case tmachine.TypeWorker.String():
			workers[m.MachineID] = configuration.MachineConfiguration()
		case tmachine.TypeInit.String():
			c.InitMachineConfiguration = configuration.MachineConfiguration()
		default:
			return nil, fmt.Errorf("unknown machine type %s", m.MachineType)
		}
	}

	c.ControlplaneMachineConfigurations = controlplanes
	c.WorkerMachineConfigurations = workers
	c.UserConfigPatches = userPatches
	c.ClientConfiguration = pulumi.StringMap{
		ClusterResourceOutputsClientConfigurationCAKey:                secrets.ClientConfiguration.CaCertificate(),
		ClusterResourceOutputsClientConfigurationClientKey:            secrets.ClientConfiguration.ClientKey(),
		ClusterResourceOutputsClientConfigurationClientCertificateKey: secrets.ClientConfiguration.ClientCertificate(),
	}

	if err := ctx.RegisterResourceOutputs(c, pulumi.Map{
		ClusterResourceOutputsControlplaneMachineConfigurations: controlplanes,
		ClusterResourceOutputsWorkerMachineConfigurations:       workers,
		ClusterResourceOutputsInitMachineConfiguration:          c.InitMachineConfiguration,
		ClusterResourceOutputsUserConfigPatches:                 userPatches,
		ClusterResourceOutputsClientConfiguration:               secrets.ClientConfiguration,
	}); err != nil {
		return nil, err
	}

	return provider.NewConstructResult(c)
}

func configureTalosInstall(image pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.All(image).ApplyT(func(args []any) (string, error) {
		image := args[0].(string)

		talosImagePatch := v1alpha1.Config{
			MachineConfig: &v1alpha1.MachineConfig{
				MachineInstall: &v1alpha1.InstallConfig{
					InstallImage: image,
				},
			},
		}
		encoded, err := yaml.Marshal(talosImagePatch)
		if err != nil {
			return "", err
		}
		return string(encoded), nil
	}).(pulumi.StringOutput)
}

func compareContractVersionWithNotify(ctx *pulumi.Context, init pulumi.StringOutput, got pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.All(got, init).ApplyT(func(v []any) string {
		got := v[0].(string)
		init := v[1].(string)
		if got != init {
			ctx.Log.Warn(fmt.Sprintf("got contract version: %s, but use init value: %s. talosVersionContract can't be changed after creation of cluster",
				got, init,
			), nil)
		}
		return init
	}).(pulumi.StringOutput)
}
