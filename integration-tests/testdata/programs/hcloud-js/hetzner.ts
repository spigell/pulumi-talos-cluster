import * as hcloud from "@pulumi/hcloud";
import * as pulumi from "@pulumi/pulumi";
import * as forge from "node-forge";
import { Cluster, DeployedServer } from "./types";

const defaultTalosInitialVersion = "v1.10.3";
const variant = "metal";

export function Hetzner(cluster: Cluster): DeployedServer[] {
  const sshKey = new hcloud.SshKey(
    "ssh",
    {
      publicKey: generateSSHKey().then((keys) => keys.publicKey),
    },
    { ignoreChanges: ["publicKey"] },
  );

  // Create the private network
  const network = new hcloud.Network("private-network", {
    name: pulumi.interpolate`private-network-${cluster.name}`,
    ipRange: cluster.privateNetwork,
  });

  // Convert network ID to integer
  const convertedNetID = network.id.apply((id) => parseInt(id, 10));

  // Add a subnet to the private network
  new hcloud.NetworkSubnet("private-subnet", {
    networkId: convertedNetID,
    type: "server",
    networkZone: "eu-central", // Adjust based on your preferred region
    ipRange: cluster.privateSubnetwork,
  });

  const deployed: DeployedServer[] = [];

  for (const machine of cluster.machines) {
    if (!machine.hcloud) {
      throw new Error(`machine ${machine.id} is missing hcloud configuration`);
    }

    const serverArch = architectureForServer(machine.hcloud.serverType);
    const machineVariant = machine.platform ?? variant;
    const talosVersion = machine.talosInitialVersion ?? defaultTalosInitialVersion;
    const selector = `os=talos,version=${talosVersion},variant=${machineVariant},arch=${serverArch}`;
    const image = hcloud.getImage({
      withSelector: selector,
      withArchitecture: serverArch,
    });

    if (!machine.privateIP) {
      throw new Error(`machine ${machine.id} is missing privateIP`);
    }

    const serverArgs: hcloud.ServerArgs = {
      name: machine.id,
      serverType: machine.hcloud.serverType,
      image: image.then((v) => `${v.id}`),
      sshKeys: [sshKey.id],
      networks: [
        {
          networkId: convertedNetID,
          ip: machine.privateIP,
        },
      ],
    };

    if (machine.hcloud.datacenter) {
      serverArgs.datacenter = machine.hcloud.datacenter;
    } else {
      serverArgs.location = "nbg1";
    }

    // Define the server
    const server = new hcloud.Server(
      machine.id,
      serverArgs,
      { ignoreChanges: ["sshKeys"] },
    );

    deployed.push({
      id: machine.id,
      ip: server.ipv4Address,
    });
  }

  return deployed;
}

function architectureForServer(serverType: string): string {
  if (serverType.startsWith("cax") || serverType.startsWith("cpx")) {
    return "arm";
  }

  return "x86";
}

async function generateSSHKey(): Promise<{ publicKey: string }> {
  return new Promise((resolve, reject) => {
    try {
      // Generate an RSA key pair
      const keypair = forge.pki.rsa.generateKeyPair(2048);

      // Convert to PEM format
      const publicKeyForge = forge.ssh.publicKeyToOpenSSH(keypair.publicKey);

      const publicKey = `${publicKeyForge}`;

      resolve({ publicKey });
    } catch (error) {
      reject(`Error generating SSH keys: ${error}`);
    }
  });
}
