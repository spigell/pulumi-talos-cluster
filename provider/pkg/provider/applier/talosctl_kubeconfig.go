package applier

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
)

// NewKubeconfig fetches a kubeconfig via talosctl using the generated talosconfig.
func (a *Applier) NewKubeconfig(endpoints []string, nodes []string, deps []pulumi.Resource) pulumi.StringOutput {
	t := talosctl.New()
	if len(endpoints) > 0 {
		t = t.WithNodeIP(endpoints[0])
	}
	dir := generateWorkDirNameForTalosctl(a.name, "kubeconfig", "kubeconfig")

	out, err := t.RunGetCommand(a.ctx, &talosctl.Args{
		TalosConfig: a.NewTalosconfig(endpoints, nodes),
		Dir:         dir,
		CommandArgs: pulumi.String("kubeconfig -"),
		RetryCount:  2,
	}, deps)
	if err != nil {
		// Return rejected output with error to propagate failure.
		return pulumi.StringOutput{}.ApplyT(func(string) (string, error) {
			return "", fmt.Errorf("run kubeconfig: %w", err)
		}).(pulumi.StringOutput)
	}

	return out
}
