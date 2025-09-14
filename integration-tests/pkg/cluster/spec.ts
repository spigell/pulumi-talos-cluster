import { readFileSync } from "fs";
import { parse } from "yaml";

export interface Machine {
  id: string;
  type: string;
  serverType: string;
  platform: string;
  talosInitialVersion: string;
  talosImage: string;
  privateIP: string;
  datacenter: string;
  configPatches: string[];
  userdata: string;
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
