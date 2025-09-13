from dataclasses import dataclass
from typing import List
import yaml


@dataclass
class Machine:
    id: str
    type: str
    serverType: str
    platform: str
    talosInitialVersion: str
    talosImage: str
    privateIP: str
    datacenter: str


@dataclass
class Cluster:
    name: str
    privateNetwork: str
    privateSubnetwork: str
    kubernetesVersion: str
    machines: List[Machine]


def load(path: str) -> Cluster:
    with open(path, "r", encoding="utf-8") as f:
        data = yaml.safe_load(f)
    machines = [Machine(**m) for m in data.get("machines", [])]
    return Cluster(
        name=data.get("name", ""),
        privateNetwork=data.get("privateNetwork", ""),
        privateSubnetwork=data.get("privateSubnetwork", ""),
        kubernetesVersion=data.get("kubernetesVersion", ""),
        machines=machines,
    )
