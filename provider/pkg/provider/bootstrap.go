package provider

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
)

type Bootstrap struct {
	pulumi.ResourceState
	BootstrapArgs
}

func BootstrapType() string {
	return ProviderName + ":index:Bootstrap"
}

type BootstrapArgs struct {
	Node                pulumi.StringOutput    `pulumi:"node"`
	ClientConfiguration pulumi.StringMapOutput `pulumi:"clientConfiguration"`
}

func bootstrap(ctx *pulumi.Context, b *Bootstrap, name string,
	args *BootstrapArgs, inputs provider.ConstructInputs, opts ...pulumi.ResourceOption,
) (*provider.ConstructResult, error) {
	// Blit the inputs onto the arguments struct.
	if err := inputs.CopyTo(args); err != nil {
		return nil, errors.Wrap(err, "setting args")
	}

	// Register our component resource.
	if err := ctx.RegisterComponentResource(BootstrapType(), name, b, opts...); err != nil {
		return nil, err
	}

	_, err := machine.NewBootstrap(ctx, fmt.Sprintf("%s:bootstrap", name), &machine.BootstrapArgs{
		ClientConfiguration: &machine.ClientConfigurationArgs{
			CaCertificate:     args.ClientConfiguration.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationCAKey)),
			ClientKey:         args.ClientConfiguration.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationClientKey)),
			ClientCertificate: args.ClientConfiguration.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationClientCertificateKey)),
		},
		Node: args.Node,
	}, pulumi.Parent(b), pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "1m", Update: "1m"}))
	if err != nil {
		return nil, err
	}

	if err := ctx.RegisterResourceOutputs(b, pulumi.Map{}); err != nil {
		return nil, err
	}

	return provider.NewConstructResult(b)
}
