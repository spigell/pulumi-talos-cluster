#!/usr/bin/env python3
"""Convert an explicit Pulumi release version to a PEP 440 package version."""

import re
import sys


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: python-version.py VERSION", file=sys.stderr)
        return 1

    match = re.fullmatch(
        r"(\d+\.\d+\.\d+)(?:-(dev|alpha|beta|rc)(?:\.(\d+))?)?(\+[0-9A-Za-z.-]+)?",
        sys.argv[1],
    )
    if match is None:
        print(f"invalid release VERSION: {sys.argv[1]}", file=sys.stderr)
        return 1

    base, stage, number, metadata = match.groups()
    suffix = {"dev": ".dev", "alpha": "a", "beta": "b", "rc": "rc"}
    print(base + (suffix[stage] + (number or "0") if stage else "") + (metadata or ""))
    return 0


if __name__ == "__main__":
    sys.exit(main())
