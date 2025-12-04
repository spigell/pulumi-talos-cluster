package talos

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	cloud "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud/go"
	taloscluster "github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster"
)

// Cluster provides helpers for creating and applying a Talos cluster
// for integration tests.
type Cluster struct {
	ctx     *pulumi.Context
	Cluster *taloscluster.Cluster

	Name     string
	machines []MachineSpec
}

// Spec describes a Talos cluster to be created.
type Spec struct {
	Name              string
	KubernetesVersion string `yaml:"kubernetesVersion"`

	Machines []MachineSpec
}

// MachineSpec contains the subset of machine fields required to create
// Talos resources.
type MachineSpec struct {
	ID            string
	Type          string
	TalosImage    string
	Platform      string
	Variant string
	SkipInitApply bool
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

		machines = append(machines, &taloscluster.ClusterMachinesArgs{
			MachineId:     server.ID(),
			NodeIp:        server.IP(),
			MachineType:   taloscluster.MachineTypes(m.Type),
			TalosImage:    pulumi.String(m.TalosImage),
			ConfigPatches: configPatches,
		})
	}

	created, err := taloscluster.NewCluster(ctx, spec.Name, &taloscluster.ClusterArgs{
		ClusterEndpoint:   pulumi.Sprintf("https://%s:6443", servers[0].IP()),
		ClusterName:       spec.Name,
		ClusterMachines:   machines,
		KubernetesVersion: pulumi.String(spec.KubernetesVersion),
	})
	if err != nil {
		return nil, fmt.Errorf("error init cluster: %w", err)
	}

	return &Cluster{
		ctx:      ctx,
		Name:     spec.Name,
		Cluster:  created,
		machines: spec.Machines,
	}, nil
}

// Apply runs the Talos Apply resource to bootstrap the cluster.
func (t *Cluster) Apply(deps []pulumi.Resource) (*Credentials, error) {
	for _, m := range t.machines {
		if m.Variant == "metal" && m.SkipInitApply {
			return nil, fmt.Errorf("skipInitApply is not supported for metal variant because reboot is required")
		}
	}

	apply, err := taloscluster.NewApply(t.ctx, t.Name, &taloscluster.ApplyArgs{
		SkipInitApply:       pulumi.Bool(t.machines[0].SkipInitApply),
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
