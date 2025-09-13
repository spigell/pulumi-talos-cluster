package talos

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster"
	taloscluster "github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster"
	"gopkg.in/yaml.v3"
)

// Cluster provides helpers for creating and applying a Talos cluster
// for integration tests.
type Cluster struct {
	ctx     *pulumi.Context
	Cluster *taloscluster.Cluster

	Name string
}

// Applied contains the credentials produced by talos.Apply.
type Applied struct {
	Kubeconfig  pulumi.StringOutput
	Talosconfig pulumi.StringOutput
}

// NewCluster creates a Talos cluster resource configured from the provided
// cluster specification and server definitions.
func NewCluster(ctx *pulumi.Context, clu *cluster.Cluster, servers []cloud.Server) (*Cluster, error) {
	if len(clu.Machines) == 0 || clu.Machines[0].Type != "init" {
		return nil, fmt.Errorf("the first node must be init")
	}

	machines := make(taloscluster.ClusterMachinesArray, 0)

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
						"validSubnets": []string{clu.PrivateNetwork},
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
					"advertisedSubnets": []string{clu.PrivateNetwork},
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

		configPatches := pulumi.StringArray{
			pulumi.String(rendered),
			pulumi.String(timePatch),
			pulumi.String(extPatch),
		}

		for _, p := range m.ConfigPatches {
			configPatches = append(configPatches, pulumi.String(p))
		}

		machines = append(machines, &taloscluster.ClusterMachinesArgs{
			MachineId:     server.ID(),
			NodeIp:        server.IP(),
			MachineType:   taloscluster.MachineTypes(m.Type),
			TalosImage:    pulumi.String(m.TalosImage),
			ConfigPatches: configPatches,
		})
	}

	created, err := taloscluster.NewCluster(ctx, clu.Name, &taloscluster.ClusterArgs{
		ClusterEndpoint: pulumi.Sprintf("https://%s:6443", servers[0].IP()),
		ClusterName:     clu.Name,
		ClusterMachines: machines,
	})
	if err != nil {
		return nil, fmt.Errorf("error init cluster: %w", err)
	}

	return &Cluster{
		ctx:     ctx,
		Name:    clu.Name,
		Cluster: created,
	}, nil
}

// Apply runs the Talos Apply resource to bootstrap the cluster.
func (t *Cluster) Apply(deps []pulumi.Resource) (*Applied, error) {
	apply, err := taloscluster.NewApply(t.ctx, t.Name, &taloscluster.ApplyArgs{
		SkipInitApply:       pulumi.Bool(true),
		ClientConfiguration: t.Cluster.ClientConfiguration,
		ApplyMachines:       t.Cluster.Machines,
	}, pulumi.DependsOn(deps), pulumi.IgnoreChanges([]string{"skipInitApply"}))
	if err != nil {
		return nil, fmt.Errorf("error apply: %w", err)
	}

	return &Applied{
		Kubeconfig:  apply.Credentials.Kubeconfig(),
		Talosconfig: apply.Credentials.Talosconfig(),
	}, nil
}
