import * as hcloud from "@pulumi/hcloud";
import * as pulumi from "@pulumi/pulumi";
import * as forge from "node-forge";
import { Cluster, DeployedServer } from "./types";

const defaultTalosInitialVersion = "v1.10.3";
const arch = "arm";
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

  const selector = `os=talos,version=${defaultTalosInitialVersion},variant=${variant},arch=${arch}`;
  const image = hcloud.getImage({
    withSelector: selector,
    withArchitecture: arch,
  });

  const deployed: DeployedServer[] = [];

  for (const machine of cluster.machines) {
    // Define the server
    const server = new hcloud.Server(
      machine.id,
      {
        name: machine.id,
        serverType: machine.serverType,
        image: image.then((v) => `${v.id}`),
        location: "nbg1",
        sshKeys: [sshKey.id],
        networks: [
          {
            networkId: convertedNetID,
            ip: machine.privateIP,
          },
        ],
      },
      { ignoreChanges: ["sshKeys"] },
    );

    deployed.push({
      id: machine.id,
      ip: server.ipv4Address,
    });
  }

  return deployed;
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
