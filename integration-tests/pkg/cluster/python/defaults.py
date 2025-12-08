from __future__ import annotations

import json
from pathlib import Path
from typing import Any, Dict, List

SCHEMA_PATH = Path(__file__).resolve().parents[1] / "schema.json"

with SCHEMA_PATH.open("r", encoding="utf-8") as f:
    _SCHEMA: Dict[str, Any] = json.load(f)


def apply_defaults(node: Any, schema: Dict[str, Any] | None = None) -> None:
    if schema is None:
        schema = _SCHEMA

    if isinstance(node, dict):
        properties = schema.get("properties", {})
        for key, prop_schema in properties.items():
            if key not in node and isinstance(prop_schema, dict) and "default" in prop_schema:
                node[key] = prop_schema["default"]
            if key in node and isinstance(prop_schema, dict):
                apply_defaults(node[key], prop_schema)
        return

    if isinstance(node, list):
        items_schema = schema.get("items")
        if isinstance(items_schema, dict):
            for item in node:
                apply_defaults(item, items_schema)
        elif isinstance(items_schema, list):
            for idx, item in enumerate(node):
                target_schema = items_schema[idx] if idx < len(items_schema) else items_schema[-1]
                if isinstance(target_schema, dict):
                    apply_defaults(item, target_schema)


def get_default(path: List[str]) -> Any:
    node: Any = _SCHEMA
    for segment in path:
        if not isinstance(node, dict) or segment not in node:
            raise KeyError(f"schema path missing segment '{segment}'")
        node = node[segment]
    if isinstance(node, dict) and "default" in node:
        return node["default"]
    return node


def schema() -> Dict[str, Any]:
    return _SCHEMA
