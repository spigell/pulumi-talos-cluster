package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster"
	talos "github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster"
	"gopkg.in/yaml.v3"
)

type Talos struct {
	ctx     *pulumi.Context
	Cluster *talos.Cluster

	Name string
}

type TalosCluster struct {
	Kubeconfig  pulumi.StringOutput
	Talosconfig pulumi.StringOutput
}

func NewTalosCluster(ctx *pulumi.Context, clu *cluster.Cluster, servers []cloud.Server) (*Talos, error) {
	if clu.Machines[0].Type != "init" {
		return nil, fmt.Errorf("the first node must be init")
	}

	machines := make(talos.ClusterMachinesArray, 0)

	for _, server := range servers {
		var m *cluster.Machine

		for _, machine := range clu.Machines {
			if machine.ID == server.ID() {
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

		timePatch, _ := yaml.Marshal(map[string]any{
			"machine": map[string]any{
				"time": map[string]any{
					"disabled": false,
				},
			},
		})

		extPatch := `apiVersion: v1alpha1
kind: ExtensionServiceConfig
name: cloudflared
environment:
  - TUNNEL_TOKEN=CHANGE_ME_AGAIN
  - TUNNEL_METRICS=localhost:2001
  - TUNNEL_EDGE_IP_VERSION=auto
`

		machines = append(machines, &talos.ClusterMachinesArgs{
			MachineId:     server.ID(),
			NodeIp:        server.IP(),
			MachineType:   talos.MachineTypes(m.Type),
			TalosImage:    pulumi.String(m.TalosImage),
			ConfigPatches: pulumi.StringArray{pulumi.String(rendered), pulumi.String(timePatch), pulumi.String(extPatch)},
		})
	}

	created, err := talos.NewCluster(ctx, clu.Name, &talos.ClusterArgs{
		ClusterEndpoint: pulumi.Sprintf("https://%s:6443", servers[0].IP()),
		ClusterName:     clu.Name,
		// KubernetesVersion: pulumi.String(clu.KubernetesVersion),
		ClusterMachines: machines,
	})
	if err != nil {
		return nil, fmt.Errorf("error init cluster: %w", err)
	}

	return &Talos{
		ctx:     ctx,
		Name:    clu.Name,
		Cluster: created,
	}, nil
}

func (t *Talos) Apply(deps []pulumi.Resource) (*TalosCluster, error) {
	apply, err := talos.NewApply(t.ctx, t.Name, &talos.ApplyArgs{
		SkipInitApply:       pulumi.Bool(true),
		ClientConfiguration: t.Cluster.ClientConfiguration,
		ApplyMachines:       t.Cluster.Machines,
	}, pulumi.DependsOn(deps), pulumi.IgnoreChanges([]string{"skipInitApply"}))
	if err != nil {
		return nil, fmt.Errorf("error apply: %w", err)
	}

	return &TalosCluster{
		Kubeconfig:  apply.Credentials.Kubeconfig(),
		Talosconfig: apply.Credentials.Talosconfig(),
	}, err
}
