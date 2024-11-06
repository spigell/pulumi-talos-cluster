package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	talos "github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster"
)

type Cluster struct {
	Name              string
	PrivateNetwork    string
	PrivateSubnetwork string
	BootTalosImageID  string
	Machines          []*ClusterMachine
}

type ClusterMachine struct {
	Name       string
	Type       talos.MachineTypes
	ServerType string
	PrivateIP  string
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cluster := &Cluster{
			Name:              ctx.Stack(),
			PrivateNetwork:    "10.10.10.0/24",
			PrivateSubnetwork: "10.10.10.0/25",
			BootTalosImageID:  "197664890",
			Machines: []*ClusterMachine{
				{
					Name:       "controlplane-1",
					Type:       talos.MachineTypesInit,
					ServerType: "cx22",
					PrivateIP:  "10.10.10.2",
				},
				{
					Name:       "worker-1",
					Type:       talos.MachineTypesWorker,
					ServerType: "cx22",
					PrivateIP:  "10.10.10.3",
				},
			},
		}

		servers, err := NewHetzner(ctx, cluster)
		if err != nil {
			return err
		}

		talosClu, err := NewTalosCluster(ctx, cluster, servers)
		if err != nil {
			return err
		}

		ctx.Export("kubeconfig", talosClu.Kubeconfig)

		return nil
	})
}
