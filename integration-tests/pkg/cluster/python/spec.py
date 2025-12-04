from dataclasses import dataclass, field
from typing import List, Optional

import yaml

from validation import validate_cluster


@dataclass
class HcloudMachine:
    serverType: str
    datacenter: Optional[str] = None


@dataclass
class Machine:
    id: str
    type: str
    platform: Optional[str] = None
    variant: str = ""
    talosInitialVersion: Optional[str] = None
    talosImage: Optional[str] = None
    privateIP: str = ""
    configPatches: List[str] = field(default_factory=list)
    userdata: Optional[str] = None
    applyConfigViaUserdata: bool = False
    hcloud: Optional[HcloudMachine] = None


@dataclass
class Cluster:
    name: str
    privateNetwork: str
    privateSubnetwork: str
    kubernetesVersion: str
    machines: List[Machine] = field(default_factory=list)
    skipInitApply: bool = False
    usePrivateNetwork: bool = False


def load(path: str) -> Cluster:
    with open(path, "r", encoding="utf-8") as f:
        data = yaml.safe_load(f) or {}
    validate_cluster(data)
    machines = [
        Machine(
            configPatches=m.get("configPatches", []),
            userdata=m.get("userdata"),
            applyConfigViaUserdata=m.get("apply-config-via-userdata", False),
            hcloud=HcloudMachine(**m["hcloud"]) if m.get("hcloud") else None,
            **{
                k: v
                for k, v in m.items()
                if k
                not in (
                    "configPatches",
                    "userdata",
                    "hcloud",
                    "apply-config-via-userdata",
                )
            }
        )
        for m in data.get("machines", [])
    ]
    return Cluster(
        name=data.get("name", ""),
        privateNetwork=data.get("privateNetwork", ""),
        privateSubnetwork=data.get("privateSubnetwork", ""),
        kubernetesVersion=data.get("kubernetesVersion", ""),
        skipInitApply=data.get("skipInitApply", False),
        usePrivateNetwork=data.get("usePrivateNetwork", False),
        machines=machines,
    )
