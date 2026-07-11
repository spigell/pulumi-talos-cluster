package applier

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
)

// GetKubeconfig fetches a kubeconfig via talosctl using the generated talosconfig.
func (a *Applier) GetKubeconfig(deps []pulumi.Resource) pulumi.StringOutput {
	t := talosctl.New().
		WithTalosConfig(a.TalosconfigForNode(a.InitNode.IP)).
		WithNodeIP(a.InitNode.IP)
	dir := generateWorkDirNameForTalosctl(a.name, "kubeconfig", "kubeconfig")

	cmd, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, "kubeconfig", a.InitNode.Name), &talosctl.Args{
		Dir:         dir,
		CommandArgs: pulumi.String("kubeconfig -"),
		RetryCount:  2,
	}, a.parent, pulumi.DependsOn(deps))
	if err != nil {
		return pulumi.StringOutput{}.ApplyT(func(string) (string, error) {
			return "", fmt.Errorf("run kubeconfig: %w", err)
		}).(pulumi.StringOutput)
	}

	return cmd.Stdout
}
