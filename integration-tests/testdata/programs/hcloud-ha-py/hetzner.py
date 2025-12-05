from __future__ import annotations

from typing import Any, Dict, List

import pulumi
from pulumi_hcloud import (
    Network,
    NetworkSubnet,
    Server,
    SshKey,
    get_image_output,
)

# Static demo key; replace with your own for real runs.
_DEMO_PUBLIC_KEY = (
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJtX7iBJ8zZCNdSP6NqBqXex12MNl81pHR38t0KBfZ1f demo@example"
)

def architecture_for_server(server_type: str) -> str:
    if server_type.startswith("cax") or server_type.startswith("cpx"):
        return "arm"
    return "x86"


def hetzner(cluster: Any) -> List[Dict[str, pulumi.Output]]:
    ssh_key = SshKey(
        "ssh",
        public_key=_DEMO_PUBLIC_KEY,
        opts=pulumi.ResourceOptions(ignore_changes=["publicKey"]),
    )

    network = Network(
        "private-network",
        name=pulumi.Output.concat("private-network-", cluster.name),
        ip_range=cluster.privateNetwork,
    )
    converted_net_id = network.id.apply(lambda net_id: int(net_id))

    NetworkSubnet(
        "private-subnet",
        network_id=converted_net_id,
        type="server",
        network_zone="eu-central",
        ip_range=cluster.privateSubnetwork,
    )

    deployed: List[Dict[str, pulumi.Output]] = []

    for machine in cluster.machines:
        if machine.hcloud is None:
            raise ValueError(f"machine {machine.id} is missing hcloud configuration")
        if not machine.privateIP:
            raise ValueError(f"machine {machine.id} is missing privateIP")

        server_arch = architecture_for_server(machine.hcloud.serverType)
        machine_variant = machine.variant
        talos_version = machine.talosInitialVersion
        selector = f"os=talos,version={talos_version},variant={machine_variant},arch={server_arch}"

        image = get_image_output(with_selector=selector, with_architecture=server_arch)

        server_args: Dict[str, Any] = {
            "name": machine.id,
            "server_type": machine.hcloud.serverType,
            "image": image.id.apply(lambda v: str(v)),
            "ssh_keys": [ssh_key.id],
            "networks": [
                {
                    "network_id": converted_net_id,
                    "ip": machine.privateIP,
                }
            ],
        }

        if machine.hcloud.datacenter:
            server_args["datacenter"] = machine.hcloud.datacenter
        else:
            server_args["location"] = "nbg1"

        server = Server(
            resource_name=machine.id,
            opts=pulumi.ResourceOptions(ignore_changes=["sshKeys", "userData"]),
            **server_args,
        )

        deployed.append({
            "id": machine.id,
            "ip": server.ipv4_address,
        })

    return deployed
