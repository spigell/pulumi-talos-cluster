import type * as pulumi from "@pulumi/pulumi";

export type HcloudMachine = {
  serverType: string;
  datacenter?: string;
};

export type Machine = {
  id: string;
  type: string;
  platform?: string;
  variant: "cloud" | "metal";
  talosInitialVersion?: string;
  talosImage?: string;
  privateIP: string;
  configPatches?: string[];
  userdata?: string;
  applyConfigViaUserdata?: boolean;
  hcloud?: HcloudMachine;
};

export type Cluster = {
  name: string;
  privateNetwork?: string;
  privateSubnetwork?: string;
  kubernetesVersion: string;
  skipInitApply?: boolean;
  machines: Machine[];
};

export type Server = {
  id(): string;
  ip(): pulumi.Output<string>;
  withUserdata(userdata: pulumi.Output<string>): Server;
};

export type Network = {
  id(): pulumi.Output<string>;
};

export type Deployed = {
  servers: Server[];
  deps: pulumi.Resource[];
};

export type DeployedCluster = {
  clusterMachines: pulumi.Output<Record<string, string>>;
  credentials: DeployedCredentials;
};

export type DeployedCredentials = {
  kubeconfig: pulumi.Output<string>;
  talosconfig: pulumi.Output<string>;
};
