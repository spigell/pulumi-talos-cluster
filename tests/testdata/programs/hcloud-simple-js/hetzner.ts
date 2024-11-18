import * as hcloud from "@pulumi/hcloud";
import * as pulumi from "@pulumi/pulumi";
import {Cluster, DeployedServer} from './types'

export function Hetzner (cluster: Cluster): DeployedServer[] {
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
        ipRange: cluster.PrivateSubnetwork,
    });

    const selector = "os=talos,testing=true"
    const image = hcloud.getImage({
        withSelector: selector,
        withArchitecture: 'arm'
    })

    const deployed: DeployedServer[] = [];

    for (const machine of cluster.machines) {
        // Define the server
        const server = new hcloud.Server(machine.id, {
            name: machine.id,
            serverType: machine.serverType,
            image: image.then(v => `${v.id}`), // OS image
            location: "nbg1",               // Choose the Hetzner location
            networks: [{
                networkId: convertedNetID,
                ip: machine.privateIP,
            }],
        });

        deployed.push({
            id: machine.id,
            ip: server.ipv4Address,
        });
    }

    return deployed;
}