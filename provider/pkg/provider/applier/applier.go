package applier

import (
	"fmt"
	"os"
	"path/filepath"

	//"github.com/pulumi/pulumi-command/sdk/go/command/local"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/client"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/hooks"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

type Applier struct {
	ctx                 *pulumi.Context
	name                string
	clientConfiguration *machine.ClientConfigurationArgs
	parent              pulumi.ResourceOption
	commnanInterpreter  pulumi.StringArray
	skipInitNode        bool

	etcdMembers   int
	etcdReadyHook *pulumi.ResourceHook

	InitNode *InitNode
}

type InitNode struct {
	IP   string
	Name string
}

func New(ctx *pulumi.Context, name string, client *machine.ClientConfigurationArgs, parent pulumi.ResourceOption) (*Applier, error) {
	a := &Applier{
		name:                name,
		ctx:                 ctx,
		parent:              parent,
		clientConfiguration: client,
		// 1 is default value, because we have at least one init node.
		etcdMembers: 1,
		commnanInterpreter: pulumi.StringArray{
			pulumi.String("/bin/bash"),
			pulumi.String("-c"),
		},
	}

	etcdReadyHook, err := a.ctx.RegisterResourceHook("health-check", hooks.EtcdReadyHook(a.ctx.Log), nil)
	if err != nil {
		return a, err
	}

	a.etcdReadyHook = etcdReadyHook

	return a, nil
}

func (a *Applier) WithSkipedInitApply(skip bool) *Applier {
	a.skipInitNode = skip

	return a
}

func (a *Applier) WithEtcdMembersCount(count int) *Applier {
	a.etcdMembers = count

	return a
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

func (a *Applier) BootstrapInitNode(m *types.MachineInfo) ([]pulumi.Resource, error) {
	// The Init node is special. We need to init by ourselves.
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

	cli, err := a.cliApply(m, tmachine.TypeInit, deps)
	if err != nil {
		return deps, err
	}

	return append(deps, cli...), nil
}

func (a *Applier) InitControlplane(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	if !a.skipInitNode {
		applied, err := a.initApply(m, deps)
		if err != nil {
			return nil, err
		}
		deps = append(deps, applied)
	}

	return deps, nil
}

func (a *Applier) ApplyToControlplane(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	cli, err := a.cliApply(m, tmachine.TypeControlPlane, deps)
	if err != nil {
		return deps, err
	}

	return append(deps, cli...), nil
}

func (a *Applier) ApplyToWorker(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	if !a.skipInitNode {
		applied, err := a.initApply(m, deps)
		if err != nil {
			return nil, err
		}
		deps = append(deps, applied)
	}

	cli, err := a.cliApply(m, tmachine.TypeWorker, deps)
	if err != nil {
		return deps, err
	}

	return append(deps, cli...), nil
}

func (a *Applier) UpgradeK8S(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	upgraded, err := a.upgradeK8S(m, deps)
	if err != nil {
		return deps, err
	}

	return append(deps, upgraded), nil
}


func (a *Applier) cliApply(m *types.MachineInfo, role tmachine.Type, deps []pulumi.Resource, ) ([]pulumi.Resource, error) {
	upgraded, err := a.upgrade(m, role, deps)
	if err != nil {
		return nil, err
	}

	deps = append(deps, upgraded)

	apply, err := a.apply(m, deps)
	if err != nil {
		return nil, err
	}

	deps = append(deps, apply)

	return deps, nil
}


func (a *Applier) initApply(m *types.MachineInfo, deps []pulumi.Resource) (pulumi.Resource, error) {

	apply, err := machine.NewConfigurationApply(a.ctx, fmt.Sprintf("%s:initial-apply:%s", a.name, m.MachineID), &machine.ConfigurationApplyArgs{
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
	if err != nil {
		return nil, err
	}

	deps = append(deps, apply)

	return a.reboot(m, deps)
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


func generateWorkDirNameForTalosctl(stack, step, machineID string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("talos-home-for-%s", stack), fmt.Sprintf("%s-%s", step, machineID))
}
