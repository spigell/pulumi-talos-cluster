import * as pulumi from "@pulumi/pulumi";
import * as talos from "@spigell/pulumi-talos-cluster"
import {Cluster} from './types'
import {Hetzner} from './hetzner'

// ARM precreated image
const imageID = '197664791'
const cluster: Cluster = {
	name: pulumi.getStack(),
	kubernetesVersion: 'v1.31.0',
	privateNetwork: '10.10.0.0/16',
	PrivateSubnetwork: '10.10.10.0/25',
	machines: [
		{
			id: 'controlplane-1',
			type: talos.MachineTypes.Init,
			bootTalosImageID: imageID,
			serverType: 'cax21',
			privateIP: '10.10.10.2'
		}
	]
}

const servers = Hetzner(cluster)
const machines: pulumi.Input<pulumi.Input<talos.types.input.ClusterMachinesArgs>[]> = []

cluster.machines.forEach(v => machines.push({
	machineId: v.id,
	nodeIp: servers.find(m => v.id == m.id)?.ip as pulumi.Input<string>,
	machineType: v.type
}))

export const clu = new talos.Cluster(cluster.name, {
	kubernetesVersion: cluster.kubernetesVersion,
	clusterEndpoint: pulumi.interpolate `https://${servers[0].ip}:6443`,
	clusterName: cluster.name,
	clusterMachines: machines,
})


export const apply = new talos.Apply(cluster.name, {
		clientConfiguration: clu.clientConfiguration,
		applyMachines: clu.machines
})
