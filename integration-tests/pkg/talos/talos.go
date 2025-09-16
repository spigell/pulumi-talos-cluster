package talos

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud"
	taloscluster "github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster"
)

// Cluster provides helpers for creating and applying a Talos cluster
// for integration tests.
type Cluster struct {
	ctx     *pulumi.Context
	Cluster *taloscluster.Cluster

	Name     string
	machines []*cluster.Machine
}

// Spec describes a Talos cluster to be created.
type Spec struct {
	Name     string
	Machines []MachineSpec
}

// MachineSpec contains the subset of machine fields required to create
// Talos resources.
type MachineSpec struct {
	ID            string
	Type          string
	TalosImage    string
	ConfigPatches []string
}

// Applied contains the credentials produced by talos.Apply.
type Credentials struct {
	Kubeconfig  pulumi.StringOutput
	Talosconfig pulumi.StringOutput
}

// NewCluster creates a Talos cluster resource configured from the provided
// cluster specification and server definitions.
func NewCluster(ctx *pulumi.Context, spec *Spec, servers []cloud.Server) (*Cluster, error) {
	if len(spec.Machines) == 0 || spec.Machines[0].Type != "init" {
		return nil, fmt.Errorf("the first node must be init")
	}

	machines := make(taloscluster.ClusterMachinesArray, 0)

	for _, server := range servers {
		var m *MachineSpec

		for i := range spec.Machines {
			if spec.Machines[i].ID == server.ID() {
				m = &spec.Machines[i]
				break
			}
		}

		configPatches := pulumi.StringArray{}
		for _, p := range m.ConfigPatches {
			configPatches = append(configPatches, pulumi.String(p))
		}

		talosImage := m.TalosImage
		if talosImage == "" {
			talosImage = "ghcr.io/siderolabs/installer:v1.10.5"
		}

		machines = append(machines, &taloscluster.ClusterMachinesArgs{
			MachineId:     server.ID(),
			NodeIp:        server.IP(),
			MachineType:   taloscluster.MachineTypes(m.Type),
			TalosImage:    pulumi.StringPtr(talosImage),
			ConfigPatches: configPatches,
		})
	}

	k8sVersion := clu.KubernetesVersion
	if k8sVersion == "" {
		k8sVersion = "v1.31.0"
	}

	created, err := taloscluster.NewCluster(ctx, clu.Name, &taloscluster.ClusterArgs{
		ClusterEndpoint:      pulumi.Sprintf("https://%s:6443", servers[0].IP()),
		ClusterName:          clu.Name,
		ClusterMachines:      machines,
		KubernetesVersion:    pulumi.StringPtr(k8sVersion),
		TalosVersionContract: pulumi.StringPtr("v1.10.5"),
	})
	if err != nil {
		return nil, fmt.Errorf("error init cluster: %w", err)
	}

	return &Cluster{
		ctx:      ctx,
		Name:     clu.Name,
		Cluster:  created,
		machines: clu.Machines,

	}, nil
}

// Apply runs the Talos Apply resource to bootstrap the cluster.
func (t *Cluster) Apply(deps []pulumi.Resource, skipInitApply bool) (*Applied, error) {
	if skipInitApply {
		for _, m := range t.machines {
			if m.Platform != "metal" {
				return nil, fmt.Errorf("skipInitApply is only supported for metal platform")
			}
		}
	}


	apply, err := taloscluster.NewApply(t.ctx, t.Name, &taloscluster.ApplyArgs{
		SkipInitApply:       pulumi.BoolPtr(skipInitApply),
		ClientConfiguration: t.Cluster.ClientConfiguration,
		ApplyMachines:       t.Cluster.Machines,
	}, pulumi.DependsOn(deps), pulumi.IgnoreChanges([]string{"skipInitApply"}))
	if err != nil {
		return nil, fmt.Errorf("error apply: %w", err)
	}

	return &Credentials{
		Kubeconfig:  apply.Credentials.Kubeconfig(),
		Talosconfig: apply.Credentials.Talosconfig(),
	}, nil
}
