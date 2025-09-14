package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	hcloud "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud/hcloud"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster"
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

		talosClu, applied, err := cluster.Deploy(ctx, clu, provider, true)
		if err != nil {
			return err
		}

		ctx.Export("clusterMachineConfigs", talosClu.Cluster.GeneratedConfigurations)
		ctx.Export("kubeconfig", applied.Kubeconfig)
		ctx.Export("talosconfig", applied.Talosconfig)

		return nil
	})
}
