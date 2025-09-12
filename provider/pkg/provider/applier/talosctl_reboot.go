package applier

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

// talosctlFastReboot returns command talosctl command for rebooting node.
// It doesn't wait for good node status.
func (a *Applier) reboot(m *types.MachineInfo, deps []pulumi.Resource) (pulumi.Resource, error) {
	stageName := "cli-reboot"
	home := generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID)
	t := talosctl.New().WithNodeIP(m.NodeIP)

	return t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		TalosConfig: a.basicClient().TalosConfig(),
		PrepareDeps: deps,
		Dir:         home,
		CommandArgs: pulumi.String(talosctlFastRebootArgs()),
		RetryCount:  0,
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "5m", Update: "5m"}),
		pulumi.DependsOn(deps),
	}...,
	)
}

func talosctlFastRebootArgs() string {
	// Do not wait for succesfull reboot.
	return strings.Join([]string{
		"reboot --wait --debug --timeout=20s",
		// Talosctl exit with code 1 if timeout exceeded.
		// This code is allowed.
		"[ $? == 1 ] && true",
		// Do not retry this command.
	}, " ; ")
}
