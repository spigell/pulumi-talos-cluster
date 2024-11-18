package applier

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/client"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

type Applier struct {
	ctx                 *pulumi.Context
	name                string
	clientConfiguration *machine.ClientConfigurationArgs
	parent              pulumi.ResourceOption
	commnanInterpreter  pulumi.StringArray

	InitNode *InitNode
}

type InitNode struct {
	IP   string
	Name string
}

func New(ctx *pulumi.Context, name string, client *machine.ClientConfigurationArgs, parent pulumi.ResourceOption) *Applier {
	return &Applier{
		name:                name,
		ctx:                 ctx,
		parent:              parent,
		clientConfiguration: client,
		commnanInterpreter: pulumi.StringArray{
			pulumi.String("/bin/bash"),
			pulumi.String("-c"),
		},
	}
}

func (a *Applier) Init(m *types.MachineInfo) ([]pulumi.Resource, error) {
	applied, err := a.initApply(m, nil)
	if err != nil {
		return nil, err
	}

	deps := []pulumi.Resource{applied}

	bootstrap, err := machine.NewBootstrap(a.ctx, fmt.Sprintf("%s:bootstrap:%s", a.name, m.MachineID), &machine.BootstrapArgs{
		ClientConfiguration: a.clientConfiguration,
		Node:                pulumi.String(m.NodeIP),
	}, a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "1m", Update: "1m"}),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}

	deps = append(deps, bootstrap)

	cli, err := a.cliApply(m, deps)
	if err != nil {
		return deps, err
	}

	return append(deps, cli...), nil
}

func (a *Applier) ApplyTo(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	inited, err := a.initApply(m, deps)
	if err != nil {
		return deps, fmt.Errorf("failed to unmarshal config from string: %w", err)
	}

	deps = append(deps, inited)

	cli, err := a.cliApply(m, deps)
	if err != nil {
		return deps, err
	}

	return append(deps, cli...), nil
}

func (a *Applier) cliApply(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	set, err := local.NewCommand(a.ctx, fmt.Sprintf("%s:cli-set-talos-version:%s", a.name, m.MachineID), &local.CommandArgs{
		Create: a.talosctlUpgradeCMD(m),
		Triggers: pulumi.Array{
			pulumi.String(m.TalosImage),
		},
		Interpreter: a.commnanInterpreter,
	},
		a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m", Update: "10m"}),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}
	deps = append(deps, set)

	cmd := a.talosctlApplyCMD(m)
	apply, err := local.NewCommand(a.ctx, fmt.Sprintf("%s:cli-apply:%s", a.name, m.MachineID), &local.CommandArgs{
		Create: cmd,
		Triggers: pulumi.Array{
			pulumi.String(m.UserConfigPatches),
			pulumi.String(m.ClusterEnpoint),
		},
		Interpreter: a.commnanInterpreter,
	}, a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "90s", Update: "90s"}),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return deps, err
	}

	deps = append(deps, apply)

	return deps, nil
}

func (a *Applier) UpgradeK8S(ma []*types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	k8s, err := local.NewCommand(a.ctx, fmt.Sprintf("%s:cli-set-k8s-version:%s", a.name, ma[0].MachineID), &local.CommandArgs{
		Create:      a.talosctlUpgradeK8SCMD(ma),
		Interpreter: a.commnanInterpreter,
		Triggers: pulumi.Array{
			pulumi.String(ma[0].KubernetesVersion),
		},
	}, a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "20m", Update: "20m"}),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return deps, err
	}

	return append(deps, k8s), nil
}

func (a *Applier) initApply(m *types.MachineInfo, deps []pulumi.Resource) (pulumi.Resource, error) {
	return machine.NewConfigurationApply(a.ctx, fmt.Sprintf("%s:initial-apply:%s", a.name, m.MachineID), &machine.ConfigurationApplyArgs{
		Node:                      pulumi.String(m.NodeIP),
		MachineConfigurationInput: pulumi.String(m.Configuration),
		// Staged is not supported in maintenance.
		// NoReboot can lead to failures.
		ApplyMode: pulumi.String("reboot"),
		OnDestroy: &machine.ConfigurationApplyOnDestroyArgs{
			Graceful: pulumi.Bool(true),
			Reboot:   pulumi.Bool(false),
			Reset:    pulumi.Bool(false),
		},
		Timeouts: &machine.TimeoutArgs{
			Create: pulumi.String("1m"),
			Update: pulumi.String("1m"),
		},
		ClientConfiguration: a.clientConfiguration,
	}, a.parent,
		// Ignore changes to machineConfigurationInput to prevent unnecessary updates since there will be an additional apply via cli.
		// Generated configuration has a contract with immutable talos version and sometimes new options can be skipped.
		pulumi.IgnoreChanges([]string{"machineConfigurationInput", "applyMode"}),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "1m", Update: "1m"}),
		pulumi.DependsOn(deps),
	)
}

func (a *Applier) basicClient() client.GetConfigurationResultOutput {
	return client.GetConfigurationOutput(a.ctx, client.GetConfigurationOutputArgs{
		ClusterName: pulumi.String(a.name),
		ClientConfiguration: &client.GetConfigurationClientConfigurationArgs{
			CaCertificate:     a.clientConfiguration.CaCertificate,
			ClientKey:         a.clientConfiguration.ClientKey,
			ClientCertificate: a.clientConfiguration.ClientCertificate,
		},
	})
}

func (a *Applier) NewTalosconfig(endpoints []string, nodes []string) client.GetConfigurationResultOutput {
	return client.GetConfigurationOutput(a.ctx, client.GetConfigurationOutputArgs{
		ClusterName: pulumi.String(a.name),
		Endpoints:   pulumi.ToStringArray(endpoints),
		Nodes:       pulumi.ToStringArray(nodes),
		ClientConfiguration: &client.GetConfigurationClientConfigurationArgs{
			CaCertificate:     a.clientConfiguration.CaCertificate,
			ClientKey:         a.clientConfiguration.ClientKey,
			ClientCertificate: a.clientConfiguration.ClientCertificate,
		},
	})
}
