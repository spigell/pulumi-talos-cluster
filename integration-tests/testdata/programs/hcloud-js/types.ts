import * as talos from "@spigell/pulumi-talos-cluster";
import * as pulumi from "@pulumi/pulumi";

export type Cluster = {
  name: string;
  kubernetesVersion: string;
  privateNetwork: string;
  privateSubnetwork: string;
  machines: ClusterMachine[];
};

export type ClusterMachine = {
  id: string;
  type: talos.MachineTypes;
  serverType: string;
  privateIP: string;
  talosImage?: string;
  datacenter?: string;
  platform?: string;
  talosInitialVersion?: string;
  configPatches?: string[];
};

export type DeployedServer = {
  id: string;
  ip: pulumi.Output<string>;
};
