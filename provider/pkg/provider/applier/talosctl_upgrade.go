package applier

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"gopkg.in/yaml.v3"
)

func (a *Applier) upgrade(m *types.MachineInfo, deps []pulumi.Resource, role string) (pulumi.Resource, error) {
	stageName := "cli-talos-upgrade"
	home := generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID)
	t := talosctl.New(a.ctx, home, deps)

	opts := []pulumi.ResourceOption{
		a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m", Update: "10m"}),
		pulumi.DependsOn(deps),
	}

	etcdMemberTarget := a.etcdMembers

	if role == "init" {
		etcdMemberTarget = 1
	}

	if role == "controlplane" || role == "init" {
		hooks := []*pulumi.ResourceHook{a.etcdReadyHook}
		opts = append(opts, pulumi.ResourceHooks(&pulumi.ResourceHookBinding{
			BeforeCreate: hooks,
			BeforeUpdate: hooks,
		}))
	}

	upgrade, err := t.RunCommand(fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.TalosctlArgs{
		PrepareConfig: a.basicClient().TalosConfig(),
		Args: talosctlUpgradeArgs(m),
		RetryCount: 10,
		Environment: pulumi.StringMap{
			"NODE_IP":            pulumi.String(m.NodeIP),
			"TALOSCTL_HOME":      pulumi.String(t.Home.Dir),
			"ETCD_MEMBER_TARGET": pulumi.String(fmt.Sprint(etcdMemberTarget)),
		},
		Triggers:    pulumi.Array{pulumi.String(m.TalosImage)},
		Interpreter: a.commnanInterpreter,
	}, opts...)

	if err != nil {
		return nil, err
	}

	return upgrade, nil
}


func talosctlUpgradeArgs(m *types.MachineInfo) pulumi.StringOutput {
	return pulumi.All(
		pulumi.String(m.NodeIP),        // string
		pulumi.String(m.Configuration), // string (YAML)
	).ApplyT(func(args []any) (string, error) {
		ip := args[0].(string)
		machineConfig := args[1].(string)

		var cfg v1alpha1.Config
		if err := yaml.Unmarshal([]byte(machineConfig), &cfg); err != nil {
			return "", fmt.Errorf("failed to unmarshal machine config: %w", err)
		}

		img := cfg.MachineConfig.Install().Image()

		base := fmt.Sprintf("upgrade --debug -n %s -e %s --image %s",
			ip, ip, img,
		)

		return base, nil
	}).(pulumi.StringOutput)
}
