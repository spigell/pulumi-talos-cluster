package provider

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"

	// "github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1".
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier"
)

var ApplyResourceKubeconfigKey = "kubeconfig"

type Apply struct {
	pulumi.ResourceState
	ApplyArgs
}

func ApplyType() string {
	return ProviderName + ":index:Apply"
}

type ApplyArgs struct {
	ClientConfiguration pulumi.StringMapOutput `pulumi:"clientConfiguration"`
	ApplyMachines       *applier.ApplyMachines `pulumi:"applyMachines"`
}

func apply(ctx *pulumi.Context, a *Apply, name string,
	args *ApplyArgs, inputs provider.ConstructInputs, opts ...pulumi.ResourceOption,
) (*provider.ConstructResult, error) {
	// Blit the inputs onto the arguments struct.
	if err := inputs.CopyTo(args); err != nil {
		return nil, errors.Wrap(err, "setting args")
	}

	applier := applier.New(ctx, name,
		buildClientConfigurationFromMap(args.ClientConfiguration),
		pulumi.Parent(a))

	// Register our component resource.
	if err := ctx.RegisterComponentResource(ApplyType(), name, a, opts...); err != nil {
		return nil, err
	}

	inited, err := applier.Init(args.ApplyMachines.Init)
	if err != nil {
		return nil, err
	}

	for _, m := range args.ApplyMachines.Controlplanes {
		applied, err := applier.ApplyTo(m, inited)
		if err != nil {
			return nil, err
		}
		inited = append(inited, applied...)
	}

	for _, m := range args.ApplyMachines.Workers {
		applied, err := applier.ApplyTo(m, inited)
		if err != nil {
			return nil, err
		}
		inited = append(inited, applied...)
	}

	if err := ctx.RegisterResourceOutputs(a, pulumi.Map{
		// ApplyResourceKubeconfigKey: kube,
	}); err != nil {
		return nil, err
	}

	return provider.NewConstructResult(a)
}

// func (a *ApplyArgs) nodesIPSFor(kind string) pulumi.StringArray {
// 	nodes := make(pulumi.StringArray, 0)
// 	for _, m := range a.ApplyMachines {
// 		nodes = append(nodes, m.Configuration.ApplyT(func(v string) (pulumi.StringOutput, error) {
// 			var config v1alpha1.Config

// 			err := yaml.Unmarshal([]byte(v), &config)
// 			if err != nil {
// 				return pulumi.String("").ToStringOutput(), err
// 			}

// 			runtime.Breakpoint()

// 			if config.MachineConfig.MachineType == kind {
// 				return m.Node, nil
// 			}

// 			return pulumi.String("").ToStringOutput(), nil
// 		}).(pulumi.StringOutput))
// 	}

// 	return nodes
// }

func buildClientConfigurationFromMap(client pulumi.StringMapOutput) *machine.ClientConfigurationArgs {
	return &machine.ClientConfigurationArgs{
		CaCertificate:     client.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationCAKey)),
		ClientKey:         client.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationClientKey)),
		ClientCertificate: client.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationClientCertificateKey)),
	}
}
