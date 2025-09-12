package applier

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"gopkg.in/yaml.v3"
)

func (a *Applier) upgrade(m *types.MachineInfo, role tmachine.Type, deps []pulumi.Resource) (pulumi.Resource, error) {
	opts := []pulumi.ResourceOption{
		a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m", Update: "10m"}),
		pulumi.DependsOn(deps),
	}

	etcdMemberTarget := a.etcdMembers

	if role == tmachine.TypeInit {
		etcdMemberTarget = 1
	}

	if role == tmachine.TypeInit || role == tmachine.TypeControlPlane {
		hooks := []*pulumi.ResourceHook{a.etcdReadyHook}
		opts = append(opts, pulumi.ResourceHooks(&pulumi.ResourceHookBinding{
			BeforeCreate: hooks,
			BeforeUpdate: hooks,
		}))
	}

	args, err := talosctlUpgradeArgs(m)
	if err != nil {
		return nil, err
	}

	stageName := "cli-upgrade"
	home := generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID)
	t := talosctl.New().WithNodeIP(m.NodeIP)

	return t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		TalosConfig: a.basicClient().TalosConfig(),
		PrepareDeps: deps,
		Dir:         home,
		CommandArgs: pulumi.String(args),
		RetryCount:  10,
		Environment: pulumi.StringMap{
			"NODE_IP":            pulumi.String(m.NodeIP),
			"TALOSCTL_HOME":      pulumi.String(home),
			"ETCD_MEMBER_TARGET": pulumi.String(fmt.Sprint(etcdMemberTarget)),
		},
		Triggers: pulumi.Array{pulumi.String(m.TalosImage)},
	}, opts...)
}

func talosctlUpgradeArgs(m *types.MachineInfo) (string, error) {
	machineConfig := m.Configuration

	var cfg v1alpha1.Config
	if err := yaml.Unmarshal([]byte(machineConfig), &cfg); err != nil {
		return "", fmt.Errorf("failed to unmarshal machine config: %w", err)
	}

	img := cfg.MachineConfig.Install().Image()

	base := fmt.Sprintf("upgrade --debug --image %s", img)

	return base, nil
}
