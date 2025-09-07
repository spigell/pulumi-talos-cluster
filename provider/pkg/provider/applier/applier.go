package applier

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

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
	skipInitNode        bool

	etcdMembers int

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

func (a *Applier) WithSkipedInitApply(skip bool) *Applier {
	a.skipInitNode = skip

	return a
}

func (a *Applier) WithEtcdMembersCount(count int) *Applier {
	a.etcdMembers = count

	return a
}

func (a *Applier) Init(m *types.MachineInfo) ([]pulumi.Resource, error) {
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

	cli, err := a.cliApply(m, deps)
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
	cli, err := a.cliApply(m, deps)
	if err != nil {
		return deps, err
	}

	return append(deps, cli...), nil
}

func (a *Applier) ApplyTo(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	if !a.skipInitNode {
		applied, err := a.initApply(m, deps)
		if err != nil {
			return nil, err
		}
		deps = append(deps, applied)
	}

	cli, err := a.cliApply(m, deps)
	if err != nil {
		return deps, err
	}

	return append(deps, cli...), nil
}

func (a *Applier) cliApply(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	etcdReadyHook, err := a.ctx.RegisterResourceHook("health-check", a.etcdReady, nil)
	if err != nil {
		return nil, err
	}
	hooks := []*pulumi.ResourceHook{etcdReadyHook}

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
		pulumi.ResourceHooks(&pulumi.ResourceHookBinding{
			BeforeCreate: hooks,
			BeforeUpdate: hooks,
		}),
	)
	if err != nil {
		return nil, err
	}
	deps = append(deps, set)

	cmd := a.talosctlApplyCMD(m, deps)
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

func (a *Applier) etcdReady(args *pulumi.ResourceHookArgs) error {
	ip := args.NewOutputs["node"].StringValue()

	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		health := exec.Command("talosctl", "-n", ip, "-e", ip, "etcd", "health")
		if err := health.Run(); err != nil {
			fmt.Printf("Health check attempt %d failed: %v\n", i+1, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		membersCmd := exec.Command("talosctl", "-n", ip, "-e", ip, "etcd", "member", "list", "-o", "json")
		out, err := membersCmd.Output()
		if err != nil {
			fmt.Printf("Health check attempt %d failed: %v\n", i+1, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		var members struct {
			Members []any `json:"members"`
		}
		if err := json.Unmarshal(out, &members); err != nil {
			fmt.Printf("Health check attempt %d failed: %v\n", i+1, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		if len(members.Members) != a.etcdMembers {
			fmt.Printf("Health check attempt %d failed: expected %d members, got %d\n", i+1, a.etcdMembers, len(members.Members))
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("health check failed after %d attempts", maxRetries)
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

	return local.NewCommand(a.ctx, fmt.Sprintf("%s:reboot:%s", a.name, m.MachineID), &local.CommandArgs{
		Create:      a.talosctlFastReboot(m),
		Interpreter: a.commnanInterpreter,
	}, a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "5m", Update: "5m"}),
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
