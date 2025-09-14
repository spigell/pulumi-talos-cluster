package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	hcloud "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud/hcloud"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster"
	talospkg "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/talos"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		clu, err := cluster.Load("cluster.yaml")
		if err != nil {
			return err
		}
		clu.Name = ctx.Stack()

		provider, err := hcloud.NewWithIPS(ctx, clu)
		if err != nil {
			return err
		}

		servers := provider.Servers()

		talosClu, err := talospkg.NewCluster(ctx, clu, servers)
		if err != nil {
			return err
		}

		for _, s := range servers {
			s.WithUserdata(talosClu.Cluster.GeneratedConfigurations.MapIndex(
				pulumi.String(s.ID()),
			).ToStringOutput().ApplyT(func(v string) string {
				ctx.Log.Debug(fmt.Sprintf("set userdata for server %s: \n\n%s\n\n===", s.ID(), v), nil)
				return v
			}).(pulumi.StringOutput))
		}

		deployed, err := provider.Up()
		if err != nil {
			return err
		}

		applied, err := talosClu.Apply(deployed.Deps, clu.SkipInitApply)
		if err != nil {
			return err
		}

		ctx.Export("clusterMachineConfigs", talosClu.Cluster.GeneratedConfigurations)
		ctx.Export("kubeconfig", applied.Kubeconfig)
		ctx.Export("talosconfig", applied.Talosconfig)

		return nil
	})
}
