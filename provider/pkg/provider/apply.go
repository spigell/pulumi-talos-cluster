package provider

import (
	"fmt"

	"github.com/pkg/errors"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	pulumi_cluster "github.com/pulumiverse/pulumi-talos/sdk/go/talos/cluster"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"

	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

type Apply struct {
	pulumi.ResourceState
	ApplyArgs

	Credentials pulumi.StringMapOutput `pulumi:"credentials"`
}

func ApplyType() string {
	return ProviderName + ":index:Apply"
}

type ApplyArgs struct {
	ClientConfiguration pulumi.StringMapOutput `pulumi:"clientConfiguration"`
	ApplyMachines       pulumi.ArrayMapOutput  `pulumi:"applyMachines"`
	SkipInitApply       pulumi.BoolOutput      `pulumi:"skipInitApply"`
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

	a.Credentials = pulumi.All(args.ApplyMachines, args.SkipInitApply).ApplyT(func(v []any) (pulumi.StringMapOutput, error) {
		creds := make(pulumi.StringMap, 0)
		endpoints := make([]string, 0)
		nodes := make([]string, 0)
		controlplanes := make([]*types.MachineInfo, 0)

		ma := v[0].(map[string][]any)

		app := applier.New(ctx, name,
			buildClientConfigurationFromMap(args.ClientConfiguration),
			pulumi.Parent(a),
		).WithSkipedInitApply(v[1].(bool))

		init := ma[tmachine.TypeInit.String()]

		cp := ma[tmachine.TypeControlPlane.String()]

		if len(init) == 0 {
			return creds.ToStringMapOutput(), fmt.Errorf("a init node must exist")
		}

		i := types.ParseMachineInfo(init[0].(map[string]any))

		endpoints = append(endpoints, i.NodeIP)
		controlplanes = append(controlplanes, i)

		app.InitNode = &applier.InitNode{
			Name: i.MachineID,
			IP:   i.NodeIP,
		}

		app.WithEtcdMembersCount(1)

		inited, err := app.Init(i)
		if err != nil {
			return creds.ToStringMapOutput(), err
		}

		app.WithEtcdMembersCount(len(cp) + 1)

		controlplanesReady := inited

		for _, m := range cp {
			ma, ok := m.(map[string]any)
			if !ok {
				return creds.ToStringMapOutput(), fmt.Errorf("expected map[string]any, got: %T", m)
			}

			node := types.ParseMachineInfo(ma)
			controlplanes = append(controlplanes, node)

			endpoints = append(endpoints, node.NodeIP)

			i, err := app.InitControlplane(node, inited)
			if err != nil {
				return creds.ToStringMapOutput(), err
			}

			controlplanesReady = append(controlplanesReady, i...)

			applied, err := app.ApplyToControlplane(node, controlplanesReady)
			if err != nil {
				return creds.ToStringMapOutput(), err
			}

			controlplanesReady = append(controlplanesReady, applied...)
		}

		// Nodes contains all nodes, including endpoints
		nodes = append(nodes, endpoints...)

		workers := ma[tmachine.TypeWorker.String()]
		for _, m := range workers {
			ma, ok := m.(map[string]any)
			if !ok {
				return creds.ToStringMapOutput(), fmt.Errorf("expected map[string]any, got: %T", m)
			}
			node := types.ParseMachineInfo(ma)

			nodes = append(nodes, node.NodeIP)

			_, err := app.ApplyTo(node, inited)
			if err != nil {
				return creds.ToStringMapOutput(), err
			}
		}

		upgraded, err := app.UpgradeK8S(controlplanes, controlplanesReady)
		if err != nil {
			return creds.ToStringMapOutput(), err
		}

		kubeconfig, err := pulumi_cluster.NewKubeconfig(ctx, types.KubeconfigKey, &pulumi_cluster.KubeconfigArgs{
			Node: pulumi.String(i.NodeIP),
			ClientConfiguration: &pulumi_cluster.KubeconfigClientConfigurationArgs{
				CaCertificate:     args.ClientConfiguration.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationCAKey)),
				ClientKey:         args.ClientConfiguration.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationClientKey)),
				ClientCertificate: args.ClientConfiguration.MapIndex(pulumi.String(ClusterResourceOutputsClientConfigurationClientCertificateKey)),
			},
		}, pulumi.Parent(a),
			pulumi.DependsOn(upgraded),
		)
		if err != nil {
			return creds.ToStringMapOutput(), err
		}

		creds[types.TalosconfigKey] = app.NewTalosconfig(endpoints, nodes).TalosConfig()
		creds[types.KubeconfigKey] = kubeconfig.KubeconfigRaw

		return creds.ToStringMapOutput(), nil
	}).(pulumi.StringMapOutput)

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
