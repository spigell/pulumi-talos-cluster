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
}

export interface Cluster {
  name: string;
  privateNetwork: string;
  privateSubnetwork: string;
  kubernetesVersion: string;
  skipInitApply?: boolean;
  machines: Machine[];
}

export function load(path: string): Cluster {
  const data = readFileSync(path, "utf8");
  const spec = parse(data) as Cluster;
  if (spec.skipInitApply === undefined) {
    spec.skipInitApply = false;
  }
  return spec;
}
