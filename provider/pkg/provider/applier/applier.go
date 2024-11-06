package applier

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/client"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
)

type Applier struct {
	ctx                 *pulumi.Context
	name                string
	clientConfiguration *machine.ClientConfigurationArgs
	parent              pulumi.ResourceOption
}

type ApplyMachines struct {
	Init          *ApplyMachine   `pulumi:"init"`
	Controlplanes []*ApplyMachine `pulumi:"controlplane"`
	Workers       []*ApplyMachine `pulumi:"worker"`
}

type ApplyMachine struct {
	MachineID         string              `pulumi:"machineId"`
	Node              pulumi.StringOutput `pulumi:"node"`
	Configuration     pulumi.StringOutput `pulumi:"configuration"`
	UserConfigPatches pulumi.StringOutput `pulumi:"userConfigPatches"`
}

func New(ctx *pulumi.Context, name string, client *machine.ClientConfigurationArgs, parent pulumi.ResourceOption) *Applier {
	return &Applier{
		name:                name,
		ctx:                 ctx,
		parent:              parent,
		clientConfiguration: client,
	}
}

func (a *Applier) Init(m *ApplyMachine) ([]pulumi.Resource, error) {
	applied, err := a.initApply(m, nil)
	if err != nil {
		return nil, err
	}

	deps := []pulumi.Resource{applied}

	bootstrap, err := machine.NewBootstrap(a.ctx, fmt.Sprintf("%s:bootstrap", a.name), &machine.BootstrapArgs{
		ClientConfiguration: a.clientConfiguration,
		Node:                m.Node,
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

func (a *Applier) ApplyTo(m *ApplyMachine, deps []pulumi.Resource) ([]pulumi.Resource, error) {
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

func (a *Applier) cliApply(m *ApplyMachine, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	set, err := local.NewCommand(a.ctx, fmt.Sprintf("%s:cli-set-talos-version:%s", a.name, m.MachineID), &local.CommandArgs{
		Create: a.talosctlUpgradeCMD(m),
		Triggers: pulumi.Array{
			m.Configuration,
		},
	}, a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m", Update: "10m"}),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}
	deps = append(deps, set)

	apply, err := local.NewCommand(a.ctx, fmt.Sprintf("%s:cli-apply:%s", a.name, m.MachineID), &local.CommandArgs{
		Create: a.talosctlApplyCMD(m),
		Triggers: pulumi.Array{
			m.UserConfigPatches,
			m.Configuration,
		},
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

func (a *Applier) initApply(m *ApplyMachine, deps []pulumi.Resource) (pulumi.Resource, error) {
	return machine.NewConfigurationApply(a.ctx, fmt.Sprintf("%s:initial-apply:%s", a.name, m.MachineID), &machine.ConfigurationApplyArgs{
		Node:                      m.Node,
		MachineConfigurationInput: m.Configuration,
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
