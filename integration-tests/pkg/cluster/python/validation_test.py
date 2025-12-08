from __future__ import annotations

from pathlib import Path
from typing import Optional
import json

import pytest
import yaml
from cluster.python.validation import validate_cluster
from cluster.python.defaults import get_default

FIXTURES_DIR = Path(__file__).resolve().parents[1] / "fixtures"
SCHEMA_PATH = Path(__file__).resolve().parents[1] / "schema.json"

with open(SCHEMA_PATH, "r", encoding="utf-8") as f:
    _SCHEMA = json.load(f)

@pytest.mark.parametrize(
    ("fixture_name", "message"),
    [
        ("load-valid.yaml", None),
        ("load-minimal.yaml", None),
        ("validation-networks-present.yaml", None),
        ("validation-anchors.yaml", None),
        ("validation-missing-name.yaml", "Invalid cluster spec: 'name' is a required string"),
        ("validation-missing-machines.yaml", "Invalid cluster spec: 'machines' must be a non-empty array"),
        ("validation-empty-machines.yaml", "Invalid cluster spec: 'machines' must be a non-empty array"),
        ("validation-missing-id.yaml", "Invalid cluster spec: 'machines[0].id' is a required string"),
        ("validation-missing-type.yaml", "Invalid cluster spec: 'machines[0].type' is a required string"),
        ("validation-missing-platform.yaml", "Invalid cluster spec: 'machines[0].platform' is a required string"),
        ("validation-unsupported-platform.yaml", "Invalid cluster spec: 'machines[0].platform' must be 'hcloud'"),
        ("validation-ip-outside.yaml", "Invalid cluster spec: machine 'worker-1' privateIP '10.0.1.10' must be inside '10.0.0.0/24'"),
        ("validation-missing-networks.yaml", "when 'usePrivateNetwork' is true, both 'privateNetwork' and 'privateSubnetwork' are required"),
        ("validation-single-network.yaml", "when 'usePrivateNetwork' is true, both 'privateNetwork' and 'privateSubnetwork' are required"),
        ("validation-unknown-top.yaml", "unknown field 'extra' is not allowed"),
        ("validation-unknown-machine.yaml", "unknown field 'machines[0].unknown' is not allowed"),
    ],
)
def test_validate_cluster(fixture_name: str, message: Optional[str]) -> None:
    data = _load_fixture(fixture_name)
    if message is None:
        validate_cluster(data)
        return

    with pytest.raises(ValueError) as exc:
        validate_cluster(data)
    assert message in str(exc.value)

def _load_fixture(name: str) -> dict:
    with open(FIXTURES_DIR / name, "r", encoding="utf-8") as f:
        return yaml.safe_load(f) or {}


def test_defaults_are_applied() -> None:
    data = _load_fixture("load-defaults.yaml")

    validate_cluster(data)

    kubernetes_version = get_default(["properties", "kubernetesVersion", "default"])
    talos_image = get_default(["properties", "machines", "items", "properties", "talosImage", "default"])

    machine = data["machines"][0]
    server_type = get_default(
        ["properties", "machineDefaults", "properties", "hcloud", "properties", "serverType", "default"]
    )
    datacenter = get_default(
        ["properties", "machineDefaults", "properties", "hcloud", "properties", "datacenter", "default"]
    )

    assert data["kubernetesVersion"] == kubernetes_version
    assert machine["talosImage"] == talos_image
    assert machine["hcloud"]["serverType"] == server_type
    assert machine["hcloud"]["datacenter"] == datacenter
