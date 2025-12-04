package cluster

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	cloud "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud/go"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/talos"
)

type Deployed struct {
	ClusterMachines pulumi.StringMapOutput
	Credentials     *DeployedCredentials
}

type DeployedCredentials struct {
	Kubeconfig  pulumi.StringOutput
	Talosconfig pulumi.StringOutput
}

// Deploy provisions servers with the given provider and boots a Talos cluster.
// If patchUserdata is true, Talos-generated configuration is set as the userdata
// for each server (unless overridden in the machine spec).
func Deploy(ctx *pulumi.Context, provider cloud.Provider, cluster *Cluster) (*Deployed, error) {
	servers := provider.Servers()

	spec := &talos.Spec{Name: cluster.Name, KubernetesVersion: cluster.KubernetesVersion, Machines: make([]talos.MachineSpec, len(cluster.Machines))}
	for i, m := range cluster.Machines {
		spec.Machines[i] = talos.MachineSpec{
			ID:            m.ID,
			Type:          m.Type,
			TalosImage:    m.TalosImage,
			Platform:      m.Platform,
			SkipInitApply: cluster.SkipInitApply,

			ConfigPatches: m.ConfigPatches,
		}
	}

	talosClu, err := talos.NewCluster(ctx, spec, servers)
	if err != nil {
		return nil, err
	}

	for _, s := range servers {
		m := machineByID(cluster.Machines, s.ID())
		if m.ApplyConfigViaUserdata {
			userdata := talosClu.Cluster.GeneratedConfigurations.MapIndex(
				pulumi.String(s.ID()),
			).ToStringOutput()

			s.WithUserdata(userdata.ToStringOutput().ApplyT(func(v string) string {
				ctx.Log.Debug(fmt.Sprintf("set userdata for server %s: \n\n%s\n\n===", s.ID(), v), nil)
				return v
			}).(pulumi.StringOutput))
		}
	}

	deployed, err := provider.Up()
	if err != nil {
		return nil, err
	}

	creds, err := talosClu.Apply(deployed.Deps)
	if err != nil {
		return nil, err
	}

	return &Deployed{
		ClusterMachines: talosClu.Cluster.GeneratedConfigurations,
		Credentials: &DeployedCredentials{
			Kubeconfig:  creds.Kubeconfig,
			Talosconfig: creds.Talosconfig,
		},
	}, nil
}

func machineByID(ma []*Machine, id string) *Machine {
	for _, m := range ma {
		if m.ID == id {
			return m
		}
	}
	return nil
}
