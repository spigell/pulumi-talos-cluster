import pytest
from pathlib import Path

from spec import load


FIXTURES_DIR = Path(__file__).resolve().parents[1] / "fixtures"


def _fixture(name: str) -> str:
    return str(FIXTURES_DIR / name)


def test_load_raises_for_missing_file():
    with pytest.raises(FileNotFoundError):
        load("non-existent.yaml")


def test_load_raises_for_malformed_yaml(tmp_path: Path):
    bad_path = tmp_path / "bad.yaml"
    bad_path.write_text("name: test-cluster\n  invalid: [", encoding="utf-8")
    with pytest.raises(Exception):
        load(str(bad_path))
