package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/tests/pkg/cluster"
	"github.com/spigell/pulumi-talos-cluster/tests/pkg/hcloud"
)


var (
	clu = &cluster.Cluster{
			PrivateNetwork:    "10.10.10.0/24",
			PrivateSubnetwork: "10.10.10.0/25",
			KubernetesVersion: "v1.31.0",
			Machines: []*cluster.Machine{
				{
					ID:       "controlplane-1",
					Type:       "init",
					ServerType: "cx22",
					PrivateIP:  "10.10.10.5",
				},
				{
					ID:       "worker-1",
					Type:       "worker",
					ServerType: "cx22",
					PrivateIP:  "10.10.10.3",
				},
			},
		}
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		clu.Name = ctx.Stack()

		hetzner, err := hcloud.NewWithIPS(ctx, clu)
		if err != nil {
			return err
		}

		talosClu, err := NewTalosCluster(ctx, clu, hetzner.Servers)
		if err != nil {
			return err
		}

		for i, s := range hetzner.Servers {
			hetzner.Servers[i] = s.WithUserdata(talosClu.Cluster.GeneratedConfigurations.MapIndex(
				pulumi.String(s.ID),
			).ToStringOutput().ApplyT(func (v string) string {
				ctx.Log.Debug(fmt.Sprintf("set userdata for server %s: %s", s.ID, v), nil)
				return v
			}).(pulumi.StringOutput))
		}

		servers, err := hetzner.Up()
		if err != nil {
			return err
		}

		applied, err := talosClu.Apply(servers.Deps)
		if err != nil {
			return err
		}

		ctx.Export("kubeconfig", applied.Kubeconfig)
		ctx.Export("talosconfig", applied.Talosconfig)

		return nil
	})
}
