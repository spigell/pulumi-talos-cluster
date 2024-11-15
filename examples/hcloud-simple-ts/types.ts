import * as talos from "@spigell/pulumi-talos-cluster"
import * as pulumi from "@pulumi/pulumi";

export type Cluster = {
	name: string;
	kubernetesVersion: string;
	privateNetwork: string;
	PrivateSubnetwork: string;
	machines: ClusterMachine[];
}

export type ClusterMachine = {
	id: string
	type: talos.MachineTypes
	bootTalosImageID: string
	serverType: string
	privateIP: string
}

export type DeployedServer = {
	id: string
	ip: pulumi.Output<string>
}
