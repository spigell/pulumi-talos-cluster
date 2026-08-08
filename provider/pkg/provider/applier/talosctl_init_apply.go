package applier

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

// initApplyWithTalosctl runs the initial apply-config step using talosctl for a machine.
func (a *Applier) initApplyWithTalosctl(m *types.MachineInfo, deps []pulumi.Resource) (pulumi.Resource, error) {
	stageName := "cli-initial-apply-config"
	t := talosctl.New().
		WithNodeIPInput(m.NodeIP).
		WithTalosConfig(a.TalosconfigForNode(m.NodeIP))
	machineConfigName := "machineconfig.yaml"

	apply, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		AdditionalFiles: []talosctl.ExtraFile{
			{Name: machineConfigName, Content: m.Configuration},
		},
		CommandArgs:      pulumi.Sprintf("apply-config -f %s --mode reboot", machineConfigName),
		Dir:              generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID),
		RetryCount:       2,
		TryInsecureFirst: true,
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.IgnoreChanges([]string{ignoreCreateChange}),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: timeoutShort, Update: timeoutShort}),
		pulumi.DependsOn(deps),
	}...)
	if err != nil {
		return nil, err
	}

	return apply, nil
}
