import { readFileSync } from "fs";
import { parse } from "yaml";

export interface HcloudMachine {
  serverType: string;
  datacenter?: string;
}

export interface Machine {
  id: string;
  type: string;
  platform: string;
  talosInitialVersion: string;
  talosImage: string;
  privateIP: string;
  configPatches: string[];
  userdata: string;
  hcloud?: HcloudMachine;
}

export interface Cluster {
  name: string;
  privateNetwork: string;
  privateSubnetwork: string;
  kubernetesVersion: string;
  machines: Machine[];
}

export function load(path: string): Cluster {
  const data = readFileSync(path, "utf8");
  return parse(data) as Cluster;
}
