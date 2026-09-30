"""Deterministic serializer. All public calls validate the condition JSON."""

import json

from .fields import FIELDS
from .validator import validate_conditions


class QueryRejected(ValueError):
    def __init__(self, validation):
        self.validation = validation
        super().__init__("；".join(e["message"] for e in validation["errors"]))


def _render(node):
    if node["type"] in ("and", "or"):
        joiner = " && " if node["type"] == "and" else " || "
        return "(" + joiner.join(_render(child) for child in node["conditions"]) + ")"
    value = node["value"]
    encoded = str(value).lower() if isinstance(value, bool) else json.dumps(str(value), ensure_ascii=False)
    operator = FIELDS[node["field"]]["operators"][node["operator"]]
    return f"{node['field']}{operator}{encoded}"


def generate_query(document):
    validation = validate_conditions(document)
    if not validation["valid"]:
        raise QueryRejected(validation)
    return _render(validation["normalized"]["condition"])
