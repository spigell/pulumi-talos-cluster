package applier

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

func (a *Applier) upgradeK8S(m *types.MachineInfo, deps []pulumi.Resource) (pulumi.Resource, error) {
	stageName := "cli-upgrade-k8s"
	home := generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID)
	t := talosctl.New().
		WithNodeIPInput(m.NodeIP).
		WithTalosConfig(a.TalosconfigForNode(m.NodeIP))

	return t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		PrepareDeps:    deps,
		Dir:            home,
		CommandArgs:    pulumi.Sprintf("upgrade-k8s --with-docs=false --with-examples=false --to %s", m.KubernetesVersion),
		RetryCount:     1,
		UpdateOnChange: true,
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "20m", Update: "20m"}),
		pulumi.DependsOn(deps),
	}...,
	)
}
