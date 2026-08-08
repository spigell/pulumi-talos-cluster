package applier

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/hooks"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
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

	if (role == tmachine.TypeInit || role == tmachine.TypeControlPlane) && a.opts.etcdHookEnabled {
		// Populate upgrade hooks lazily and reuse if already registered.
		if _, ok := a.hooks[hookStageUpgrade]; !ok {
			h, err := a.ctx.RegisterResourceHook("health-check", hooks.EtcdReadyHook(a.ctx.Log), nil)
			if err != nil {
				return nil, err
			}
			a.hooks[hookStageUpgrade] = []*pulumi.ResourceHook{h}
		}

		if hooksForStage, ok := a.hooks[hookStageUpgrade]; ok && len(hooksForStage) > 0 {
			opts = append(opts, pulumi.ResourceHooks(&pulumi.ResourceHookBinding{
				BeforeCreate: hooksForStage,
				BeforeUpdate: hooksForStage,
			}))
		}
	}

	stageName := "cli-upgrade"
	home := generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID)
	t := talosctl.New().
		WithNodeIPInput(m.NodeIP).
		WithTalosConfig(a.TalosconfigForNode(m.NodeIP))
	upgradeArgs := m.TalosImage.ToStringOutput().ApplyT(func(image string) (string, error) {
		return talosctlUpgradeArgs(image, m.MachineID)
	}).(pulumi.StringOutput)

	return t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		PrepareDeps: deps,
		Dir:         home,
		CommandArgs: upgradeArgs,
		RetryCount:  10,
		Environment: pulumi.StringMap{
			"NODE_IP":            m.NodeIP,
			"TALOSCTL_HOME":      pulumi.String(home),
			"ETCD_MEMBER_TARGET": pulumi.String(fmt.Sprint(etcdMemberTarget)),
		},
		UpdateOnChange: true,
	}, opts...)
}

func talosctlUpgradeArgs(image, machineID string) (string, error) {
	img := strings.TrimSpace(image)
	if img == "" {
		return "", fmt.Errorf("talos image is required for machine %s", machineID)
	}

	// --drain defaults to true since talosctl v1.13 and requires a kubeconfig
	// from the target node, which workers cannot serve during provisioning.
	base := fmt.Sprintf("upgrade --debug --drain=false --image %s", img)

	return base, nil
}
