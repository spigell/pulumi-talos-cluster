from __future__ import annotations

from pathlib import Path
from typing import Any, List, Optional
import ipaddress
import json

from jsonschema import Draft7Validator, ValidationError as SchemaValidationError

SCHEMA_PATH = Path(__file__).resolve().parents[1] / "schema.json"
with SCHEMA_PATH.open("r", encoding="utf-8") as schema_file:
    _VALIDATOR = Draft7Validator(json.load(schema_file))


def validate_cluster(data: dict[str, Any]) -> None:
    message = _first_validation_error(data)
    if message:
        raise ValueError(message) from None

    validation_cidr = ""
    if data.get("usePrivateNetwork"):
        private_network = (data.get("privateNetwork") or "").strip()
        private_subnetwork = (data.get("privateSubnetwork") or "").strip()
        if not private_network or not private_subnetwork:
            raise ValueError(
                "when 'usePrivateNetwork' is true, both 'privateNetwork' and 'privateSubnetwork' are required"
            )
        validation_cidr = private_subnetwork

    if validation_cidr:
        for machine in data.get("machines", []):
            _validate_machine(machine, validation_cidr)

def _validate_machine(machine: dict[str, Any], validation_cidr: str) -> None:
    private_ip = (machine.get("privateIP") or "").strip()
    machine_id = machine.get("id", "<unknown>")

    if not private_ip:
        raise ValueError(
            f"Invalid cluster spec: machine '{machine_id}' must define privateIP when usePrivateNetwork is true"
        )

    _assert_ip_in_network(private_ip, validation_cidr, machine_id)

def _first_validation_error(data: dict[str, Any]) -> Optional[str]:
    try:
        _VALIDATOR.validate(data)
    except SchemaValidationError as exc:
        return _format_validation_error(exc)
    return None


def _format_validation_error(exc: SchemaValidationError) -> str:
    if exc.validator == "required":
        missing = _extract_name(exc.message, default="value")
        if missing == "machines":
            return "Invalid cluster spec: 'machines' must be a non-empty array"
        path = _format_path(list(exc.absolute_path), missing)
        return f"Invalid cluster spec: '{path}' is a required string"

    if exc.validator == "additionalProperties":
        prop = _extract_name(exc.message, default="field")
        path = _format_path(list(exc.absolute_path), prop)
        return f"Invalid cluster spec: unknown field '{path}' is not allowed"

    if exc.validator == "minItems" and list(exc.absolute_path) == ["machines"]:
        return "Invalid cluster spec: 'machines' must be a non-empty array"

    if exc.validator == "enum" and list(exc.absolute_path)[-1:] == ["platform"]:
        path = _format_path(list(exc.absolute_path))
        return f"Invalid cluster spec: '{path}' must be 'hcloud'"

    path = _format_path(list(exc.absolute_path))
    if path:
        return f"Invalid cluster spec: '{path}' {exc.message}"
    return f"Invalid cluster spec: {exc.message}"


def _format_path(parts: List[Any], missing: Optional[str] = None) -> str:
    tokens: List[str] = []
    for idx, part in enumerate(parts):
        if isinstance(part, int):
            tokens.append(f"[{part}]")
        else:
            tokens.append(part if idx == 0 else f".{part}")
    if missing:
        tokens.append(missing if not tokens else f".{missing}")
    return "".join(tokens)


def _extract_name(message: str, default: str) -> str:
    if "'" in message:
        return message.split("'")[1]
    return default


def _assert_ip_in_network(ip: str, cidr: str, machine_id: str) -> None:
    network = ipaddress.ip_network(cidr, strict=False)
    address = ipaddress.ip_address(ip)
    if address not in network:
        raise ValueError(
            f"Invalid cluster spec: machine '{machine_id}' privateIP '{ip}' must be inside '{cidr}'"
        )
