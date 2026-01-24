package provider

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/siderolabs/talos/pkg/machinery/gendata"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier"
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
	ClusterResourceOutputsTalosconfig                             = "talosconfig"
)

type Cluster struct {
	pulumi.ResourceState
	types.Cluster

	ClientConfiguration     pulumi.StringMapOutput `pulumi:"clientConfiguration"`
	GeneratedConfigurations pulumi.StringMap       `pulumi:"generatedConfigurations"`
	Machines                pulumi.ArrayMap        `pulumi:"machines"`
	Talosconfig             pulumi.StringOutput    `pulumi:"talosconfig"`
}

func ClusterType() string {
	return ProviderName + ":index:Cluster"
}

// type ClusterArgs types.Cluster

func GenerateDefaultInstallerImage() string {
	return fmt.Sprintf("%s/%s/installer:%s", gendata.ImagesRegistry, gendata.ImagesUsername, gendata.VersionTag)
}

func cluster(ctx *pulumi.Context, c *Cluster, name string,
	args *types.Cluster, inputs provider.ConstructInputs, opts ...pulumi.ResourceOption,
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
		pulumi.StringMap{},
		pulumi.Parent(c),
	)
	if err != nil {
		return nil, err
	}

	// Generate secrets via talosctl.
	secrets, err := app.GenerateSecrets()
	if err != nil {
		return nil, errors.Wrap(err, "generating secrets")
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

		// cfg := talosctl.New().CatFile(ctx, filepath.Join(workDir, "configs"), cfgFile, []pulumi.Resource{configCmd})
		// 		-               configuration := machine.GetConfigurationOutput(ctx, machine.GetConfigurationOutputArgs{
		// -                       ClusterName:       pulumi.String(args.ClusterName),
		// -                       MachineType:       pulumi.String(machineType),
		// -                       ClusterEndpoint:   args.ClusterEndpoint,
		// -                       KubernetesVersion: args.KubernetesVersion,
		// -                       TalosVersion:      compareContractVersionWithNotify(ctx, secrets.TalosVersion, args.TalosVersionContract.ToStringOutput()),
		// -                       ConfigPatches: pulumi.All(
		// -                               m.ConfigPatches, // StringArrayInput  -> []string in ApplyT
		// -                               configureTalosInstall(m.TalosImage.ToStringPtrOutput().Elem()), // StringInput -> string in ApplyT
		// -                       ).ApplyT(func(args []any) []string {
		// -                               base := args[0].([]string) // from m.ConfigPatches
		// -                               extra := args[1].(string)  // from configureTalosInstall(...)
		// -                               // append safely (copy if you care about not aliasing base)
		// -                               out := make([]string, 0, len(base)+1)
		// -                               out = append(out, base...)
		// -                               out = append(out, extra)
		// -                               return out
		// -                       }).(pulumi.StringArrayOutput),
		// -
		// -                       MachineSecrets: secrets.ToSecretsOutput().MachineSecrets(),
		// -               }, nil)
		// -
		// -               generated[m.MachineID] = configuration.MachineConfiguration()

		// Generate configs via talosctl using the generated secrets.
		configuration, err := app.GenerateMachineConfig(args, m, secrets)
		if err != nil {
			return nil, errors.Wrap(err, "generating configs")
		}

		generated[m.MachineID] = configuration

		mInfo := m.ToMachineInfoMap(args.ClusterEndpoint, args.KubernetesVersion, configuration)

		switch m.MachineType {
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

	talosconfig, err := app.GenerateTalosconfig(args, secrets)
	if err != nil {
		return nil, errors.Wrap(err, "generating talosconfig")
	}

	c.ClientConfiguration = talosconfig.ApplyTWithContext(ctx.Context(), func(_ context.Context, raw string) (map[string]string, error) {
		ca, key, cert, err := applier.ExtractTalosconfigCreds(raw, args.ClusterName)
		if err != nil {
			return nil, err
		}

		return map[string]string{
			ClusterResourceOutputsClientConfigurationCAKey:                ca,
			ClusterResourceOutputsClientConfigurationClientKey:            key,
			ClusterResourceOutputsClientConfigurationClientCertificateKey: cert,
		}, nil
	}).(pulumi.StringMapOutput)
	c.Talosconfig = talosconfig

	if err := ctx.RegisterResourceOutputs(c, pulumi.Map{
		ClusterResourceOutputsClientConfiguration:     c.ClientConfiguration,
		ClusterResourceOutputsMachines:                c.Machines,
		ClusterResourceOutputsGeneratedConfigurations: generated,
		ClusterResourceOutputsTalosconfig:             c.Talosconfig,
		// "secretsStash": secretsStash.Output,
	}); err != nil {
		return nil, err
	}

	return provider.NewConstructResult(c)
}
