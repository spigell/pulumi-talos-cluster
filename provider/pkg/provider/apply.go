package provider

import (
	"fmt"

	"github.com/pkg/errors"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"

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
	ClientConfiguration pulumi.StringMapInput      `pulumi:"clientConfiguration"`
	ApplyMachines       pulumi.MapArrayMapOutput   `pulumi:"applyMachines"`
	MachineTopology     pulumi.StringArrayMapInput `pulumi:"machineTopology"`
	SkipInitApply       pulumi.BoolOutput          `pulumi:"skipInitApply"`
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

	// machineTopology is optional so programs written for SDKs without it keep working.
	topology := args.MachineTopology
	if topology == nil {
		topology = machineTopologyFromApplyMachines(args.ApplyMachines)
	}

	a.Credentials = pulumi.All(topology, args.SkipInitApply).ApplyT(func(v []any) (pulumi.StringMapOutput, error) {
		creds := make(pulumi.StringMap, 0)
		endpoints := make(pulumi.StringArray, 0)
		nodes := make(pulumi.StringArray, 0)

		ma := v[0].(map[string][]string)

		init := ma[tmachine.TypeInit.String()]
		if len(init) == 0 {
			return creds.ToStringMapOutput(), fmt.Errorf("a init node must exist")
		}
		cp := ma[tmachine.TypeControlPlane.String()]
		workers := ma[tmachine.TypeWorker.String()]

		cfg := args.ClientConfiguration

		app, err := applier.New(ctx, name, cfg, pulumi.Parent(a))
		if err != nil {
			return creds.ToStringMapOutput(), err
		}

		app.WithHooks(true)
		app.WithSkipedInitApply(v[1].(bool))
		app.WithEtcdMembersCount(len(cp) + 1)

		i := machineInfoAt(args.ApplyMachines, tmachine.TypeInit.String(), 0, init[0])

		endpoints = append(endpoints, i.NodeIP)

		app.InitNode = &applier.InitNode{
			Name: i.MachineID,
			IP:   i.NodeIP,
		}

		inited, err := app.BootstrapInitNode(i)
		if err != nil {
			return creds.ToStringMapOutput(), err
		}

		controlplanesReady := inited

		for index, machineID := range cp {
			node := machineInfoAt(args.ApplyMachines, tmachine.TypeControlPlane.String(), index, machineID)

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

		for index, machineID := range workers {
			node := machineInfoAt(args.ApplyMachines, tmachine.TypeWorker.String(), index, machineID)

			nodes = append(nodes, node.NodeIP)

			workerDeps, err := app.ApplyToWorker(node, inited)
			if err != nil {
				return creds.ToStringMapOutput(), err
			}

			controlplanesReady = append(controlplanesReady, workerDeps...)
		}

		controlplanesReady, err = app.UpgradeK8S(i, controlplanesReady)
		if err != nil {
			return creds.ToStringMapOutput(), err
		}

		creds[types.TalosconfigKey] = pulumi.ToSecret(app.Talosconfig(endpoints, nodes)).(pulumi.StringOutput)
		// We use only one endpoint for kubeconfig because talosctl doesn't support multiple endpoints for this command.
		// It is safe because we can fetch kubeconfig from any node.
		creds[types.KubeconfigKey] = pulumi.ToSecret(app.GetKubeconfig(controlplanesReady)).(pulumi.StringOutput)

		return creds.ToStringMapOutput(), nil
	}).(pulumi.StringMapOutput)

	if err := ctx.RegisterResourceOutputs(a, pulumi.Map{
		types.KubeconfigKey:  a.Credentials.MapIndex(pulumi.String(types.KubeconfigKey)),
		types.TalosconfigKey: a.Credentials.MapIndex(pulumi.String(types.TalosconfigKey)),
	}); err != nil {
		return nil, err
	}

	return provider.NewConstructResult(a)
}

// machineTopologyFromApplyMachines derives plain machine IDs grouped by type from applyMachines.
// Unlike an explicit machineTopology, it resolves only when applyMachines is known.
func machineTopologyFromApplyMachines(machines pulumi.MapArrayMapOutput) pulumi.StringArrayMapOutput {
	return pulumi.Unsecret(machines.ApplyT(func(byType map[string][]map[string]any) (map[string][]string, error) {
		topology := make(map[string][]string, len(byType))
		for role, list := range byType {
			ids := make([]string, 0, len(list))
			for _, m := range list {
				id, ok := m[types.MachineIDKey].(string)
				if !ok {
					return nil, fmt.Errorf("machine of type %s has no %s", role, types.MachineIDKey)
				}
				ids = append(ids, id)
			}
			topology[role] = ids
		}

		return topology, nil
	})).(pulumi.StringArrayMapOutput)
}

func machineInfoAt(machines pulumi.MapArrayMapOutput, role string, index int, machineID string) *types.MachineInfo {
	machine := machines.MapIndex(pulumi.String(role)).Index(pulumi.Int(index))
	stringValue := func(key string) pulumi.StringOutput {
		return machine.MapIndex(pulumi.String(key)).ApplyT(func(value any) string {
			if value == nil {
				return ""
			}

			return value.(string)
		}).(pulumi.StringOutput)
	}
	publicStringValue := func(key string) pulumi.StringOutput {
		return pulumi.Unsecret(stringValue(key)).(pulumi.StringOutput)
	}

	return &types.MachineInfo{
		MachineID:         machineID,
		NodeIP:            publicStringValue(types.NodeIPKey),
		ClusterEnpoint:    publicStringValue(types.ClusterEnpointKey),
		UserConfigPatches: stringValue(types.UserConfigPatchesKey),
		TalosImage:        publicStringValue(types.TalosImageKey),
		KubernetesVersion: publicStringValue(types.KubernetesVersionKey),
		Configuration:     stringValue(types.ConfigurationKey),
	}
}
