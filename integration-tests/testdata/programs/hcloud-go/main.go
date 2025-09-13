package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud"
	hcloud "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud/hcloud"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster"
	talospkg "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/talos"
)

var (
	platform   = "hcloud"
	talosImage = "ghcr.io/siderolabs/installer:v1.11.0"
)

var clu = &cluster.Cluster{
	PrivateNetwork:    "10.10.10.0/24",
	PrivateSubnetwork: "10.10.10.0/25",
	KubernetesVersion: "1.32.0",
	Machines: []*cluster.Machine{
		{
			ID:         "controlplane-1",
			Type:       "init",
			Platform:   platform,
			TalosImage: talosImage,
			ServerType: "cx22",
			PrivateIP:  "10.10.10.5",
			Datacenter: "fsn1-dc14",
		},
		{
			ID:         "worker-1",
			Type:       "worker",
			Platform:   platform,
			TalosImage: talosImage,
			ServerType: "cx22",
			PrivateIP:  "10.10.10.3",
			Datacenter: "fsn1-dc14",
		},
	},
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		clu.Name = ctx.Stack()

		var (
			provider cloud.Provider
			err      error
		)
		provider, err = hcloud.NewWithIPS(ctx, clu)
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

		applied, err := talosClu.Apply(deployed.Deps)
		if err != nil {
			return err
		}

		ctx.Export("clusterMachineConfigs", talosClu.Cluster.GeneratedConfigurations)
		ctx.Export("kubeconfig", applied.Kubeconfig)
		ctx.Export("talosconfig", applied.Talosconfig)

		return nil
	})
}
