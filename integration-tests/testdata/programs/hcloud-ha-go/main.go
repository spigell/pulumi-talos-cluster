package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/hcloud"
	talos "github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster"
	"gopkg.in/yaml.v3"
)

var (
	platform   = "metal"
	talosImage = "ghcr.io/siderolabs/installer:v1.10.3"
)

var clu = &cluster.Cluster{
	PrivateNetwork:    "10.10.10.0/24",
	PrivateSubnetwork: "10.10.10.0/25",
	KubernetesVersion: "v1.31.0",
	Machines: []*cluster.Machine{
		{
			ID:         "controlplane-1",
			Type:       "init",
			TalosImage: talosImage,
			Platform:   platform,
			ServerType: "cx22",
			PrivateIP:  "10.10.10.5",
		},
		{
			ID:         "controlplane-2",
			Type:       string(talos.MachineTypesControlplane),
			Platform:   platform,
			TalosImage: talosImage,

			ServerType: "cx22",
			PrivateIP:  "10.10.10.2",
			//	Datacenter: "fsn1-dc14",
		},
		{
			ID:         "controlplane-3",
			Type:       string(talos.MachineTypesControlplane),
			ServerType: "cx22",
			Platform:   platform,
			TalosImage: talosImage,
			PrivateIP:  "10.10.10.10",
			//	Datacenter: "fsn1-dc14",
		},
		{
			ID:         "worker-1",
			Type:       "worker",
			Platform:   platform,
			TalosImage: talosImage,
			ServerType: "cx22",
			PrivateIP:  "10.10.10.3",
			//	Datacenter: "fsn1-dc14",
		},
	},
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		clu.Name = ctx.Stack()
		if clu.Machines[0].Type != "init" {
			return fmt.Errorf("the first node must be init")
		}

		machines := make(talos.ClusterMachinesArray, 0)

		servers, err := hcloud.NewWithIPS(ctx, clu)
		if err != nil {
			return err
		}

		up, err := servers.Up()
		if err != nil {
			return err
		}

		for _, server := range up.Servers {
			var m *cluster.Machine

			for _, machine := range clu.Machines {
				if machine.ID == server.ID {
					m = machine
					break
				}
			}

			patches := map[string]any{
				"debug": false,
				"machine": map[string]any{
					"kubelet": map[string]any{
						"nodeIP": map[string]any{
							"validSubnets": []string{
								clu.PrivateNetwork,
							},
						},
					},
					"time": map[string]any{
						"disabled": true,
					},
				},
			}

			if m.Type == "controlplane" || m.Type == "init" {
				patches["cluster"] = map[string]any{
					"etcd": map[string]any{
						"advertisedSubnets": []string{
							clu.PrivateNetwork,
						},
					},
				}
			}

			rendered, _ := yaml.Marshal(patches)

			machines = append(machines, &talos.ClusterMachinesArgs{
				MachineId:     server.ID,
				NodeIp:        server.IP,
				MachineType:   talos.MachineTypes(m.Type),
				TalosImage:    pulumi.String(m.TalosImage),
				ConfigPatches: pulumi.StringArray{pulumi.String(rendered)},
			})
		}

		created, err := talos.NewCluster(ctx, clu.Name, &talos.ClusterArgs{
			ClusterEndpoint:   pulumi.Sprintf("https://%s:6443", up.Servers[0].IP),
			ClusterName:       clu.Name,
			KubernetesVersion: pulumi.String(clu.KubernetesVersion),
			ClusterMachines:   machines,
		}, pulumi.DependsOn(up.Deps))
		if err != nil {
			return fmt.Errorf("error init cluster: %w", err)
		}

		apply, err := talos.NewApply(ctx, clu.Name, &talos.ApplyArgs{
			ClientConfiguration: created.ClientConfiguration,
			ApplyMachines:       created.Machines,
		})
		if err != nil {
			return fmt.Errorf("error apply: %w", err)
		}

		ctx.Export("kubeconfig", apply.Credentials.Kubeconfig())
		ctx.Export("talosconfig", apply.Credentials.Talosconfig())

		return nil
	})
}
