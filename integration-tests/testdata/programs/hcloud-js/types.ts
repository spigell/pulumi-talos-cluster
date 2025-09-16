import * as talos from "@spigell/pulumi-talos-cluster";
import * as pulumi from "@pulumi/pulumi";

export type Cluster = {
  name: string;
  kubernetesVersion: string;
  privateNetwork: string;
  privateSubnetwork: string;
  machines: ClusterMachine[];
};

export type HcloudMachine = {
  serverType: string;
  datacenter?: string;
};

export type ClusterMachine = {
  id: string;
  type: talos.MachineTypes;
  privateIP: string;
  talosImage?: string;
  platform?: string;
  talosInitialVersion?: string;
  configPatches?: string[];
  userdata?: string;
  hcloud?: HcloudMachine;
};

export type DeployedServer = {
  id: string;
  ip: pulumi.Output<string>;
};
