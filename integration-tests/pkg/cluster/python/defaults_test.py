from __future__ import annotations

from cluster.python.defaults import get_default, schema
from cluster.python.validation import validate_cluster
import yaml
from pathlib import Path


def test_schema_defaults_present() -> None:
    assert "properties" in schema()
    assert get_default(["properties", "kubernetesVersion", "default"]) is not None
    assert (
        get_default(["properties", "machines", "items", "properties", "talosImage", "default"])
        is not None
    )
    assert (
        get_default(
            ["properties", "machineDefaults", "properties", "hcloud", "properties", "serverType", "default"]
        )
        is not None
    )
    assert (
        get_default(
            ["properties", "machineDefaults", "properties", "hcloud", "properties", "datacenter", "default"]
        )
        is not None
    )


def test_defaults_are_applied() -> None:
    fixture_path = Path(__file__).resolve().parents[1] / "fixtures" / "load-defaults.yaml"
    with open(fixture_path, "r", encoding="utf-8") as f:
        data = yaml.safe_load(f) or {}

    validate_cluster(data)

    assert data["kubernetesVersion"] == get_default(["properties", "kubernetesVersion", "default"])

    machine = data["machines"][0]
    assert (
        machine["hcloud"]["serverType"]
        == get_default(
            ["properties", "machineDefaults", "properties", "hcloud", "properties", "serverType", "default"]
        )
    )
    assert (
        machine["hcloud"]["datacenter"]
        == get_default(
            ["properties", "machineDefaults", "properties", "hcloud", "properties", "datacenter", "default"]
        )
    )
    assert (
        machine["talosImage"]
        == get_default(["properties", "machines", "items", "properties", "talosImage", "default"])
    )
