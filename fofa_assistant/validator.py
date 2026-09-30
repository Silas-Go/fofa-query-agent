"""Validate an untrusted condition document without producing query text."""

import ipaddress
import re

from .fields import FIELDS

INTENT_ERRORS = {
    "unsupported_intent": ("UNSUPPORTED_INTENT", "暂未支持", "当前解析器无法完整表达这段意图，不能省略它继续生成。"),
    "ambiguous_intent": ("AMBIGUOUS_INTENT", "需要澄清", "请给出“最安全”等描述的具体可检索标准。"),
    "unobservable_intent": ("UNEXPRESSIBLE_INTENT", "能力不支持", "缺少容器监控数据或运行时访问权限，资产搜索无法确定实时内存使用率。"),
    "invalid_structure": ("INVALID_STRUCTURE", "输入非法", "括号或引号不完整，无法确定条件分组。"),
    "invalid_input": ("INVALID_INPUT", "输入非法", "自然语言输入必须为1–20000字符的非空文本。"),
    "too_complex": ("UNSUPPORTED_COMPLEXITY", "暂未支持", "条件嵌套超过当前处理上限。"),
}


def validate_conditions(document):
    errors = []
    count = 0

    def error(code, message, path, status="输入非法"):
        errors.append({"code": code, "path": path, "message": message, "status": status})

    def visit(node, path, depth=0):
        nonlocal count
        count += 1
        if count > 256 or depth > 16:
            error("UNSUPPORTED_COMPLEXITY", "最多支持256个节点、16层嵌套。", path, "暂未支持")
            return None
        if not isinstance(node, dict):
            error("INVALID_STRUCTURE", "条件节点必须是对象。", path)
            return None
        kind = node.get("type")
        if kind == "unresolved":
            if set(node) != {"type", "text", "reason"} or not isinstance(node.get("text"), str) or not isinstance(node.get("reason"), str) or node["reason"] not in INTENT_ERRORS:
                error("INVALID_STRUCTURE", "未解析意图节点格式错误。", path)
            else:
                code, status, message = INTENT_ERRORS[node["reason"]]
                error(code, message, path, status)
            return None
        if kind in ("and", "or"):
            children = node.get("conditions")
            if set(node) != {"type", "conditions"} or not isinstance(children, list) or not 2 <= len(children) <= 64:
                error("INVALID_STRUCTURE", "组合节点必须包含2–64个子条件，且不能携带其他字段。", path)
                return None
            return {"type": kind, "conditions": [visit(c, f"{path}.conditions[{i}]", depth + 1) for i, c in enumerate(children)]}
        if kind != "predicate" or set(node) != {"type", "field", "operator", "value"}:
            error("INVALID_STRUCTURE", "只接受规范条件节点，不接受查询字符串或额外属性。", path)
            return None
        name, op, value = node["field"], node["operator"], node["value"]
        if not isinstance(name, str) or name not in FIELDS:
            error("UNSUPPORTED_FIELD", f"当前字段白名单不支持：{name!r}", path + ".field", "暂未支持")
            return None
        rule = FIELDS[name]
        if not isinstance(op, str) or op not in rule["operators"]:
            error("UNSUPPORTED_OPERATOR", f"字段 {name} 不支持操作 {op!r}。", path + ".operator", "暂未支持")
            return None
        try:
            value_kind = rule["kind"]
            if value_kind == "integer":
                if type(value) is not int or not rule["minimum"] <= value <= rule["maximum"]:
                    raise ValueError(f"{name} 必须是 {rule['minimum']}–{rule['maximum']} 范围内的整数。")
            elif value_kind == "boolean":
                if type(value) is not bool:
                    raise ValueError(f"{name} 必须是布尔值 true 或 false。")
            else:
                if not isinstance(value, str) or not value or len(value) > 20000:
                    raise ValueError(f"{name} 必须是非空字符串，长度不超过20000。")
                if value_kind == "ip":
                    value = str(ipaddress.ip_network(value, strict=False)) if "/" in value else str(ipaddress.ip_address(value))
                elif value_kind == "domain":
                    if len(value) > 253 or not re.fullmatch(r"(?=.{1,253}$)[A-Za-z0-9](?:[A-Za-z0-9.-]*[A-Za-z0-9])?\.[A-Za-z]{2,}", value) or any(not re.fullmatch(r"[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?", label) for label in value.split(".")):
                        raise ValueError("domain 不是当前支持的有效域名格式。")
                    value = value.lower()
                elif value_kind == "enum" and value not in rule["values"]:
                    raise ValueError(f"{name} 的值不在当前支持的枚举范围：{value}")
        except ValueError as exc:
            error("INVALID_VALUE", str(exc), path + ".value")
            return None
        return {"type": "predicate", "field": name, "operator": op, "value": value}

    normalized = None
    if not isinstance(document, dict) or set(document) != {"version", "condition"} or document.get("version") != "1.0":
        error("INVALID_STRUCTURE", "需要 version=1.0 和 condition 两个属性。", "$")
    else:
        root = visit(document["condition"], "$.condition")
        if not errors:
            normalized = {"version": "1.0", "condition": root}
            try:
                if all(branch_conflict(branch) for branch in branches(root)):
                    error("CONFLICT", "所有可选分支都存在无法同时满足的条件。", "$.condition", "条件矛盾")
            except OverflowError:
                error("UNSUPPORTED_COMPLEXITY", "组合展开超过128个分支，请减少条件。", "$.condition", "暂未支持")
    return {"valid": not errors, "errors": errors, "normalized": normalized if not errors else None}


def branches(node):
    """Bounded conjunction branches; a dead OR branch does not reject live ones."""
    if node["type"] == "predicate":
        return [[node]]
    result = [] if node["type"] == "or" else [[]]
    for child in node["conditions"]:
        options = branches(child)
        if node["type"] == "or":
            if len(result) + len(options) > 128:
                raise OverflowError
            result.extend(options)
        else:
            if len(result) * len(options) > 128:
                raise OverflowError
            result = [a + b for a in result for b in options]
    return result


def branch_conflict(leaves):
    by_field = {}
    for leaf in leaves:
        by_field.setdefault(leaf["field"], []).append(leaf)
    for name, predicates in by_field.items():
        if name == "ip":
            continue
        equal = [p["value"] for p in predicates if p["operator"] == "eq"]
        unequal = [p["value"] for p in predicates if p["operator"] == "ne"]
        if len(set(equal)) > 1 or any(v in unequal for v in equal):
            return True
        required = [p["value"] for p in predicates if p["operator"] == "contains"]
        forbidden = [p["value"] for p in predicates if p["operator"] == "not_contains"]
        if any(v in forbidden for v in required + equal):
            return True
    return ip_conflict(by_field)


def ip_conflict(by_field):
    predicates = by_field.get("ip", [])
    if not predicates:
        return False
    includes = [ipaddress.ip_network(p["value"], strict=False) for p in predicates if p["operator"] == "eq"]
    excludes = [ipaddress.ip_network(p["value"], strict=False) for p in predicates if p["operator"] == "ne"]
    version_flags = [p["value"] for p in by_field.get("is_ipv6", [])]
    for version, bits in ((4, 32), (6, 128)):
        if any(n.version != version for n in includes) or any(flag != (version == 6) for flag in version_flags):
            continue
        low = max((int(n.network_address) for n in includes), default=0)
        high = min((int(n.broadcast_address) for n in includes), default=(1 << bits) - 1)
        if low > high:
            continue
        # Subtract excluded intervals from the remaining intersection.
        cursor = low
        for start, end in sorted((int(n.network_address), int(n.broadcast_address)) for n in excludes if n.version == version):
            if end < cursor:
                continue
            if start > cursor:
                break
            cursor = max(cursor, end + 1)
        if cursor <= high:
            return False
    return True


def result_status(validation):
    if validation["valid"]:
        return "成功"
    statuses = {e["status"] for e in validation["errors"]}
    return next(s for s in ("输入非法", "条件矛盾", "需要澄清", "能力不支持", "暂未支持") if s in statuses)
