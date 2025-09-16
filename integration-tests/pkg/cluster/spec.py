from dataclasses import dataclass
from typing import List, Optional
import yaml


@dataclass
class HcloudMachine:
    serverType: str
    datacenter: Optional[str] = None


@dataclass
class Machine:
    id: str
    type: str
    platform: str
    talosInitialVersion: str
    talosImage: str
    privateIP: str
    configPatches: List[str]
    userdata: str
    hcloud: Optional[HcloudMachine] = None


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
    machines = [
        Machine(
            configPatches=m.get("configPatches", []),
            userdata=m.get("userdata", ""),
            hcloud=HcloudMachine(**m["hcloud"]) if m.get("hcloud") else None,
            **{
                k: v
                for k, v in m.items()
                if k not in ("configPatches", "userdata", "hcloud")
            }
        )
        for m in data.get("machines", [])
    ]
    return Cluster(
        name=data.get("name", ""),
        privateNetwork=data.get("privateNetwork", ""),
        privateSubnetwork=data.get("privateSubnetwork", ""),
        kubernetesVersion=data.get("kubernetesVersion", ""),
        machines=machines,
    )
