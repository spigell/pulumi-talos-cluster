package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	talos "github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster"
	"gopkg.in/yaml.v3"
)

type TalosCluster struct {
	Kubeconfig pulumi.StringOutput
}

func NewTalosCluster(ctx *pulumi.Context, cluster *Cluster, servers []*DeployedServer) (*TalosCluster, error) {
	if cluster.Machines[0].Type != talos.MachineTypesInit {
		return nil, fmt.Errorf("the first node must be init")
	}

	machines := make(talos.ClusterMachinesArray, 0)

	for _, server := range servers {
		var m *ClusterMachine

		for _, machine := range cluster.Machines {
			if machine.Name == server.Name {
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
							cluster.PrivateNetwork,
						},
					},
				},
				"time": map[string]any{
					"disabled": false,
				},
			},
		}

		if m.Type == talos.MachineTypesInit || m.Type == talos.MachineTypesControlplane {
			patches["cluster"] = map[string]any{
				"etcd": map[string]any{
					"advertisedSubnets": []string{
							cluster.PrivateNetwork,
					},
				},
			}
		}

		rendered, _ := yaml.Marshal(patches)

		machines = append(machines, &talos.ClusterMachinesArgs{
			MachineId:  server.Name,
			NodeIp: server.IP,
			TalosImage: pulumi.String("factory.talos.dev/installer/9bf23bf8cf3fc88b4eacdd5370d613237508ca5627ce3b70900ffb15e26c9e70:v1.8.2"),
			MachineType:   m.Type,
			ConfigPatches: pulumi.String(rendered),
		})
	}

	clu, err := talos.NewCluster(ctx, cluster.Name, &talos.ClusterArgs{
		//ClusterEndpoint: pulumi.String(fmt.Sprintf("https://%s:6443", cluster.Machines[0].PrivateIP)),
		ClusterEndpoint: pulumi.Sprintf("https://%s:6443", servers[0].IP),
		ClusterName:     cluster.Name,
		KubernetesVersion: pulumi.String(cluster.KubernetesVersion),
		// TalosVersionContract: pulumi.String("v1.8.1"),
		ClusterMachines: machines,
	})
	if err != nil {
		return nil, fmt.Errorf("error init cluster: %w", err)
	}


	apply, err := talos.NewApply(ctx, cluster.Name, &talos.ApplyArgs{
		ClientConfiguration: clu.ClientConfiguration,
		ApplyMachines: clu.Machines,
	})
	if err != nil {
		return nil, fmt.Errorf("error apply: %w", err)
	}

	ctx.Export("cluster", clu)
	ctx.Export("apply", apply)
	ctx.Export("config", clu.ClientConfiguration)

	return &TalosCluster{}, err
}
