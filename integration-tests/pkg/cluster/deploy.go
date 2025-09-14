package cluster

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud"
	talospkg "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/talos"
)

// Deploy provisions servers with the given provider and boots a Talos cluster.
// If patchUserdata is true, Talos-generated configuration is set as the userdata
// for each server (unless overridden in the machine spec).
func Deploy(ctx *pulumi.Context, clu *Cluster, provider cloud.Provider, patchUserdata bool) (*talospkg.Cluster, *talospkg.Applied, error) {
	servers := provider.Servers()

	spec := &talospkg.Spec{Name: clu.Name, Machines: make([]talospkg.MachineSpec, len(clu.Machines))}
	for i, m := range clu.Machines {
		spec.Machines[i] = talospkg.MachineSpec{
			ID:            m.ID,
			Type:          m.Type,
			TalosImage:    m.TalosImage,
			ConfigPatches: m.ConfigPatches,
		}
	}

	talosClu, err := talospkg.NewCluster(ctx, spec, servers)
	if err != nil {
		return nil, nil, err
	}

	if patchUserdata {
		for _, s := range servers {
			m := machineByID(clu, s.ID())
			var userdata pulumi.StringInput
			if m != nil && m.Userdata != "" {
				userdata = pulumi.String(m.Userdata)
			} else {
				userdata = talosClu.Cluster.GeneratedConfigurations.MapIndex(
					pulumi.String(s.ID()),
				).ToStringOutput()
			}
			s.WithUserdata(userdata.ToStringOutput().ApplyT(func(v string) string {
				ctx.Log.Debug(fmt.Sprintf("set userdata for server %s: \n\n%s\n\n===", s.ID(), v), nil)
				return v
			}).(pulumi.StringOutput))
		}
	}

	deployed, err := provider.Up()
	if err != nil {
		return nil, nil, err
	}

	applied, err := talosClu.Apply(deployed.Deps)
	if err != nil {
		return nil, nil, err
	}

	return talosClu, applied, nil
}

func machineByID(c *Cluster, id string) *Machine {
	for _, m := range c.Machines {
		if m.ID == id {
			return m
		}
	}
	return nil
}
