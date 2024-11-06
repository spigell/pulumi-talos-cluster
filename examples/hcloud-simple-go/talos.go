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
	workers := make(talos.ApplyMachinesArray, 0)
	controlplanes := make(talos.ApplyMachinesArray, 0)
	var init talos.ApplyMachinesArgs
	ips := make(map[string]pulumi.StringOutput)

	for _, server := range servers {
		var m *ClusterMachine

		for _, machine := range cluster.Machines {
			if machine.Name == server.Name {
				m = machine
				break
			}
		}

		patches, _ := yaml.Marshal(map[string]any{
			"debug": true,
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
		})

		machines = append(machines, &talos.ClusterMachinesArgs{
			MachineId:  server.Name,
			TalosImage: pulumi.String("factory.talos.dev/installer/9bf23bf8cf3fc88b4eacdd5370d613237508ca5627ce3b70900ffb15e26c9e70:v1.8.2"),
			// KubernetesVersion: pulumi.String("v1.30.0"),
			MachineType:   m.Type,
			ConfigPatches: pulumi.String(patches),
		})

		ips[server.Name] = server.IP
	}

	clu, err := talos.NewCluster(ctx, cluster.Name, &talos.ClusterArgs{
		ClusterEndpoint: pulumi.String(fmt.Sprintf("https://%s:6443", cluster.Machines[0].PrivateIP)),
		ClusterName:     cluster.Name,
		// TalosVersionContract: pulumi.String("v1.8.1"),
		ClusterMachines: machines,
	})
	if err != nil {
		return nil, fmt.Errorf("error init cluster: %w", err)
	}

	for _, server := range servers {
		var m *ClusterMachine
		for _, machine := range cluster.Machines {
			if machine.Name == server.Name {
				m = machine
				break
			}
		}
		var configuration pulumi.StringOutput
		switch m.Type {
		case talos.MachineTypesControlplane:
			configuration = clu.ControlplaneMachineConfigurations.MapIndex(pulumi.String(server.Name))
			controlplanes = append(controlplanes, &talos.ApplyMachinesArgs{
				Node:              ips[server.Name],
				MachineId:         pulumi.String(server.Name),
				Configuration:     configuration,
				UserConfigPatches: clu.UserConfigPatches.MapIndex(pulumi.String(server.Name)),
			})
		case talos.MachineTypesWorker:
			configuration = clu.WorkerMachineConfigurations.MapIndex(pulumi.String(server.Name))
			workers = append(workers, &talos.ApplyMachinesArgs{
				Node:              ips[server.Name],
				MachineId:         pulumi.String(server.Name),
				Configuration:     configuration,
				UserConfigPatches: clu.UserConfigPatches.MapIndex(pulumi.String(server.Name)),
			})
		case talos.MachineTypesInit:
			configuration = clu.InitMachineConfiguration.ToStringPtrOutput().Elem()
			init = talos.ApplyMachinesArgs{
				Node:              ips[server.Name],
				MachineId:         pulumi.String(server.Name),
				Configuration:     configuration,
				UserConfigPatches: clu.UserConfigPatches.MapIndex(pulumi.String(server.Name)),
			}
		}
	}

	_, err = talos.NewApply(ctx, cluster.Name, &talos.ApplyArgs{
		ClientConfiguration: clu.ClientConfiguration.Elem().ToClientConfigurationOutput(),
		ApplyMachines: &talos.ApplyMachinesByTypeArgs{
			Init:         init,
			Worker:       workers,
			Controlplane: controlplanes,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error apply: %w", err)
	}

	// bootstraped, err := talos.NewBootstrap(ctx, cluster.Name, &talos.BootstrapArgs{
	//	Node:                servers[0].IP,
	//	ClientConfiguration: clu.ClientConfiguration.Elem().ToClientConfigurationOutput(),
	// }, pulumi.DependsOn([]pulumi.Resource{apply}))

	// if err != nil {
	//	return nil, fmt.Errorf("error bootstrap: %w", err)
	//}

	///tx.Export("boo", bootstraped)
	ctx.Export("clu", clu)
	ctx.Export("clu2", clu.ClientConfiguration)

	return &TalosCluster{}, err
}
