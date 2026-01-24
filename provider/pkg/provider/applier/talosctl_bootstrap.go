package applier

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

func (a *Applier) initApplyWithTalosctl(m *types.MachineInfo, deps []pulumi.Resource) (pulumi.Resource, error) {
	stageName := "initial-apply-config"
	t := talosctl.New().
		WithNodeIP(m.NodeIP).
		WithTalosConfig(a.TalosconfigForNode(m.NodeIP))
	machineConfigName := "machineconfig.yaml"

	apply, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		AdditionalFiles: []talosctl.ExtraFile{
			{Name: machineConfigName, Content: pulumi.String(m.Configuration)},
		},
		CommandArgs:      pulumi.Sprintf("apply-config -f %s --mode reboot", machineConfigName),
		Dir:              generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID),
		RetryCount:       2,
		TryInsecureFirst: true,
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.IgnoreChanges([]string{"create"}),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "90s", Update: "90s"}),
		pulumi.DependsOn(deps),
	}...)
	if err != nil {
		return nil, err
	}

	return apply, nil
}

//nolint:unused // kept for future use and parity with previous flow
func (a *Applier) bootstrapWithTalosctl(m *types.MachineInfo, deps []pulumi.Resource) (pulumi.Resource, error) {
	stageName := "bootstrap"
	t := talosctl.New().
		WithNodeIP(m.NodeIP).
		WithTalosConfig(a.TalosconfigForNode(m.NodeIP))

	bootstrap, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		CommandArgs: pulumi.String("bootstrap"),
		Dir:         generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID),
		RetryCount:  2,
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "90s", Update: "90s"}),
		pulumi.DependsOn(deps),
	}...)
	if err != nil {
		return nil, err
	}

	return bootstrap, nil
}
