package main

import (
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	talos "github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster"
)

type Cluster struct {
	Name              string
	PrivateNetwork    string
	PrivateSubnetwork string
	KubernetesVersion string
	Machines          []*ClusterMachine
}

type ClusterMachine struct {
	Name       string
	Type       talos.MachineTypes
	ServerType string
	PrivateIP  string
}

var (
	cluster = &Cluster{
			PrivateNetwork:    "10.10.10.0/24",
			PrivateSubnetwork: "10.10.10.0/25",
			KubernetesVersion: "v1.31.0",
			Machines: []*ClusterMachine{
				{
					Name:       "controlplane-1",
					Type:       talos.MachineTypesInit,
					ServerType: "cx22",
					PrivateIP:  "10.10.10.2",
				},
				{
					Name:       "worker-1",
					Type:       "worker",
					ServerType: "cx22",
					PrivateIP:  "10.10.10.3",
				},
			},
		}
	MockServers = []*DeployedServer{{
		IP: pulumi.String("127.0.0.1").ToStringOutput(),
		Name: "controlplane-1-mock",
		PrivateIP: pulumi.String("10.10.10.1").ToStringOutput(),
	}}
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cluster.Name = ctx.Stack()

		servers := make([]*DeployedServer, 0)
		var err error
		switch os.Getenv("USE_MOCK_SERVERS") {
		case "true":
			servers = MockServers
		default:
			servers, err = NewHetzner(ctx, cluster)
			if err != nil {
				return err
			}
		}

		talosClu, err := NewTalosCluster(ctx, cluster, servers)
		if err != nil {
			return err
		}

		ctx.Export("kubeconfig", talosClu.Kubeconfig)

		return nil
	})
}
