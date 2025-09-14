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

		provider, err := hcloud.NewWithIPS(ctx, clu)
		if err != nil {
			return err
		}

		deployed, err := cluster.Deploy(ctx, provider, clu)
		if err != nil {
			return err
		}

		ctx.Export("clusterMachineConfigs", deployed.ClusterMachines)
		ctx.Export("kubeconfig", deployed.Credentials.Kubeconfig)
		ctx.Export("talosconfig", deployed.Credentials.Talosconfig)

		return nil
	})
}
