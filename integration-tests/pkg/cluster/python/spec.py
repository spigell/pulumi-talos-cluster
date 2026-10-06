from dataclasses import dataclass, field

import yaml

from .validation import validate_cluster


@dataclass
class HcloudMachine:
    serverType: str
    location: str | None = None


@dataclass
class Machine:
    id: str
    type: str
    platform: str | None = None
    variant: str = ""
    talosInitialVersion: str | None = None
    talosImage: str | None = None
    privateIP: str = ""
    configPatches: list[str] = field(default_factory=list)
    userdata: str | None = None
    applyConfigViaUserdata: bool = False
    hcloud: HcloudMachine | None = None


@dataclass
class Cluster:
    name: str
    privateNetwork: str
    privateSubnetwork: str
    kubernetesVersion: str
    machines: list[Machine] = field(default_factory=list)
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
