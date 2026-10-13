package applier

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

type Applier struct {
	ctx                 *pulumi.Context
	name                string
	clientConfiguration pulumi.StringMapOutput
	parent              pulumi.ResourceOption
	commnanInterpreter  pulumi.StringArray
	skipInitNode        bool

	etcdMembers int
	opts        Options
	hooks       map[string][]*pulumi.ResourceHook

	InitNode *InitNode
}

type Options struct {
	etcdHookEnabled bool
}

const hookStageUpgrade = "cli-upgrade"

type InitNode struct {
	IP   pulumi.StringInput
	Name string
}

func New(ctx *pulumi.Context, name string, client pulumi.StringMapInput, parent pulumi.ResourceOption) (*Applier, error) {
	a := &Applier{
		name:                name,
		ctx:                 ctx,
		parent:              parent,
		clientConfiguration: client.ToStringMapOutput(),
		// 1 is default value, because we have at least one init node.
		etcdMembers: 1,
		commnanInterpreter: pulumi.StringArray{
			pulumi.String("/bin/bash"),
			pulumi.String("-c"),
		},
		opts: Options{etcdHookEnabled: true},
	}

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

func (a *Applier) WithHooks(enabled bool) *Applier {
	a.opts.etcdHookEnabled = enabled
	if enabled && a.hooks == nil {
		a.hooks = make(map[string][]*pulumi.ResourceHook)
	}

	return a
}

// TalosconfigForNode generates a talosconfig for a single node (IP used for endpoint and node).
func (a *Applier) TalosconfigForNode(ip pulumi.StringInput) pulumi.StringOutput {
	nodes := pulumi.StringArray{ip}
	return a.buildTalosConfig(nodes, nodes)
}

// Talosconfig generates a talosconfig for a set of endpoints and nodes.
func (a *Applier) Talosconfig(endpoints pulumi.StringArrayInput, nodes pulumi.StringArrayInput) pulumi.StringOutput {
	return a.buildTalosConfig(endpoints, nodes)
}

func (a *Applier) BootstrapInitNode(m *types.MachineInfo) ([]pulumi.Resource, error) {
	// Intentionally skip the bootstrap/reboot phase for init nodes.
	// The previous reboot-based bootstrap step was removed on purpose.
	applied, err := a.initApply(m, nil)
	if err != nil {
		return nil, err
	}

	deps := make([]pulumi.Resource, 0, 2)
	deps = append(deps, applied)

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

func (a *Applier) cliApply(m *types.MachineInfo, role tmachine.Type, deps []pulumi.Resource) ([]pulumi.Resource, error) {
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
	return a.initApplyWithTalosctl(m, deps)
}

func (a *Applier) GenerateSecrets() (pulumi.StringOutput, error) {
	return a.generateSecrets()
}

func (a *Applier) GenerateMachineConfig(c *types.Cluster, m *types.ClusterMachine, secrets pulumi.StringOutput) (pulumi.StringOutput, error) {
	configuration, err := a.generateMachineConfig(c, m, secrets)
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	return configuration.(*local.Command).Stdout, nil
}

func (a *Applier) GenerateTalosconfig(c *types.Cluster, secrets pulumi.StringOutput) (pulumi.StringOutput, error) {
	talosconfig, err := a.generateTalosconfig(c, secrets)
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	return talosconfig.(*local.Command).Stdout, nil
}

func generateWorkDirNameForTalosctl(stack, step, machineID string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("talos-home-for-%s", stack), fmt.Sprintf("%s-%s", step, machineID))
}
