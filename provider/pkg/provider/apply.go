package provider

import (
	"github.com/pkg/errors"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"

	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

var ApplyResourceKubeconfigKey = "kubeconfig"

type Apply struct {
	pulumi.ResourceState
	ApplyArgs

	O pulumi.StringOutput
}

func ApplyType() string {
	return ProviderName + ":index:Apply"
}

type ApplyArgs struct {
	ClientConfiguration pulumi.StringMapOutput `pulumi:"clientConfiguration"`
	ApplyMachines       pulumi.ArrayMapOutput  `pulumi:"applyMachines"`
}

type ApplyMachines struct {
	InitMachineConfiguration          *types.MachineInfo   `pulumi:"init"`
	ControlplaneMachineConfigurations []*types.MachineInfo `pulumi:"controlplane"`
	WorkerMachineConfigurations       []*types.MachineInfo `pulumi:"worker"`
}

func apply(ctx *pulumi.Context, a *Apply, name string,
	args *ApplyArgs, inputs provider.ConstructInputs, opts ...pulumi.ResourceOption,
) (*provider.ConstructResult, error) {
	// Blit the inputs onto the arguments struct.
	if err := inputs.CopyTo(args); err != nil {
		return nil, errors.Wrap(err, "setting args")
	}

	// Register our component resource.
	if err := ctx.RegisterComponentResource(ApplyType(), name, a, opts...); err != nil {
		return nil, err
	}

	c := pulumi.All(args.ApplyMachines).ApplyT(func(v []any) (string, error) {
		ma := v[0].(map[string][]any)

		app := applier.New(ctx, name,
			buildClientConfigurationFromMap(args.ClientConfiguration),
			pulumi.Parent(a),
		)

		init := ma[tmachine.TypeInit.String()]

		if len(init) == 0 {
			return "", nil
		}

		i := types.ParseMachineInfo(init[0].(map[string]any))
		app.InitNode = &applier.InitNode{
			Name: i.MachineID,
			IP:   i.NodeIP,
		}

		inited, err := app.Init(i)
		if err != nil {
			return "", nil
		}
		cp := ma[tmachine.TypeControlPlane.String()]
		for _, m := range cp {
			ma, ok := m.(map[string]any)
			if !ok {
				return "ERROR", nil
			}
			applied, err := app.ApplyTo(types.ParseMachineInfo(ma), inited)
			if err != nil {
				return "", nil
			}
			inited = append(inited, applied...)
		}

		workers := ma[tmachine.TypeWorker.String()]
		for _, m := range workers {
			ma, ok := m.(map[string]any)
			if !ok {
				return "ERROR", nil
			}
			applied, err := app.ApplyTo(types.ParseMachineInfo(ma), inited)
			if err != nil {
				return "", nil
			}
			inited = append(inited, applied...)
		}

		_, err = app.UpgradeK8S(types.ParseMachineInfo(init[0].(map[string]any)), inited)
		if err != nil {
			return "", nil
		}

		return "", nil
	}).(pulumi.StringOutput)

	a.O = c

	if err := ctx.RegisterResourceOutputs(a, pulumi.Map{
		// ApplyResourceKubeconfigKey: kube,
	}); err != nil {
		return nil, err
	}

	return provider.NewConstructResult(a)
}

func buildClientConfigurationFromMap(client pulumi.StringMapOutput) *machine.ClientConfigurationArgs {
	return &machine.ClientConfigurationArgs{
		CaCertificate:     client.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationCAKey)),
		ClientKey:         client.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationClientKey)),
		ClientCertificate: client.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationClientCertificateKey)),
	}
}
