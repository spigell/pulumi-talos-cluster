package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	hcloud "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud/hcloud"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster"
	talos "github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster"
	"gopkg.in/yaml.v3"
)

var (
	platform   = "metal"
	talosImage = "ghcr.io/siderolabs/installer:v1.10.6"
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

		for _, m := range clu.Machines {
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
			m.ConfigPatches = []string{string(rendered)}
		}

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
