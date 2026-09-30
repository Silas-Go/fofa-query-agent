"""Natural language to JSON conditions only; no query generation here."""

import re

from .fields import TEXT_FIELDS

COUNTRIES = {
    "中国": "CN", "美国": "US", "英国": "GB", "大不列颠及北爱尔兰联合王国": "GB",
    "德国": "DE", "日本": "JP", "韩国": "KR", "荷兰": "NL", "加拿大": "CA",
    "土耳其": "TR", "新西兰": "NZ",
}
SERVICES = {"dns": "dns", "ssh": "ssh", "ftp": "ftp", "snmp": "snmp", "memcached": "memcached"}
END = r"\s*(?:的)?(?:全部|所有)?(?:资产记录|网站类资产|资产|网站|网页|数据)?"
DOMAIN = r"[A-Za-z0-9](?:[A-Za-z0-9.-]*[A-Za-z0-9])?\.[A-Za-z]{2,}"

def pred(field, value, op="="):
    operators = {"=": "contains" if field in TEXT_FIELDS else "eq", "==": "eq", "!=": "not_contains" if field in TEXT_FIELDS else "ne"}
    return {"type": "predicate", "field": field, "operator": operators[op], "value": value}


def group(op, children):
    return children[0] if len(children) == 1 else {"type": "and" if op == "&&" else "or", "conditions": children}


def unresolved(text, reason="unsupported_intent"):
    return {"type": "unresolved", "text": text, "reason": reason}


def clean(text):
    s = text.strip().rstrip("。！？?!").strip()
    s = re.sub(r"^(?:请帮我查询|请帮我查|帮我找一下|我想要搜索|我想要查询|我想搜索|我想查询|请查询|请搜索|请帮我找|帮我查询|我想找|我想查|我想看|帮我找|搜索|查询|查找|找|查)\s*", "", s)
    return s.strip()


def country(value):
    value = value.strip()
    if value in COUNTRIES:
        return COUNTRIES[value]
    if re.fullmatch(r"[A-Z]{2}", value):
        return value
    return None


def literal_ok(value):
    # Unquoted natural-language clauses must never be swallowed as a keyword.
    return bool(value.strip()) and not re.search(r"[，,；;\n]|或者|或|并且|且|同时|但是|但|要求|任意|至少|不包含|不限定", value)


def atom(s):
    s = s.strip().rstrip("。")
    # Explicit verbatim text has priority over conjunction parsing.
    m = re.fullmatch(r"(?:网页源码|网页正文)中包含这段原样文本(?:的资产)?[：:](.+)", s, re.S)
    if m:
        return pred("body", m[1])
    m = re.fullmatch(r"字段\s*([A-Za-z_][\w.]*)\s*(?:为|等于)\s*(.+?)" + END, s)
    if m and literal_ok(m[2]):
        return {"type": "predicate", "field": m[1], "operator": "eq", "value": m[2]}
    m = re.fullmatch(r"(?:IP\s*(?:地址)?\s*(?:为|是)?\s*|地址或网段(?:为|是)\s*)([\d.]+(?:/\d+)?)" + END, s, re.I)
    if m:
        return pred("ip", m[1])
    m = re.fullmatch(r"([\d.]+/\d+)\s*网段内" + END, s)
    if m:
        return pred("ip", m[1])
    m = re.fullmatch(r"(?:这个\s*IP\s*的)?整个\s*B\s*段[：:]\s*([\d.]+)", s, re.I)
    if m:
        return pred("ip", m[1] + "/16")
    for pattern, field, op in [
        (r"主域(?:是|为)\s*(" + DOMAIN + r")", "domain", "="),
        (r"网站名(?:是|为)\s*(" + DOMAIN + r")", "host", "=="),
        (r"域名(?:里)?(?:带有|包含)\s*(\S+?)", "host", "="),
        (r"自治系统号(?:为|是)\s*(\d+)", "asn", "="),
        (r"操作系统(?:被)?识别为\s*(\S+?)", "os", "="),
        (r"网络组织名称(?:里)?包含\s*(\S+?)", "org", "="),
    ]:
        m = re.fullmatch(pattern + r"\s*" + END, s, re.I)
        if m and literal_ok(m[1]):
            return pred(field, int(m[1]) if field == "asn" else m[1].strip(), op)
    m = re.fullmatch(r"(?:开放\s*(-?\d+)\s*端口|端口(?:为|是)\s*(-?\d+))" + END, s)
    if m:
        p = int(m[1] or m[2])
        return pred("port", p)
    m = re.fullmatch(r"(?:网页)?状态码(?:为|是)\s*(\d+)" + END, s)
    if m:
        node = pred("status_code", int(m[1]))
        return group("&&", [node, pred("type", "subdomain")]) if "网站类资产" in s else node
    m = re.fullmatch(r"(?:使用\s*)?(TCP|UDP)\s*(?:协议|传输协议)" + END, s, re.I)
    if m:
        return pred("base_protocol", m[1].lower())
    m = re.fullmatch(r"(DNS|SSH|FTP|SNMP|Memcached)\s*(?:协议|服务|缓存服务协议)" + END, s, re.I)
    if m:
        return pred("protocol", SERVICES[m[1].lower()])
    m = re.fullmatch(r"(.+?)的\s*(DNS|SSH|FTP|SNMP|Memcached)\s*服务" + END, s, re.I)
    if m and country(m[1]):
        return group("&&", [pred("country", country(m[1])), pred("protocol", SERVICES[m[2].lower()])])
    m = re.fullmatch(r"(?:位于\s*|国家(?:代码)?(?:又?必须)?(?:等于|为|是)\s*|地域限定为\s*)(.+?)" + END, s)
    if m and country(m[1]):
        return pred("country", country(m[1]))
    m = re.fullmatch(r"国家(?:代码)?(?:又?必须)?(?:不等于|不是)\s*(.+?)" + END, s)
    if m and country(m[1]):
        return pred("country", country(m[1]), "!=")
    m = re.fullmatch(r"(?:不是|非)\s*IPv6" + END, s, re.I)
    if m:
        return pred("is_ipv6", False)
    m = re.fullmatch(r"(已|没有|未)绑定域名" + END, s)
    if m:
        return pred("is_domain", m[1] == "已")
    # Values must not contain additional conditions. Match whole clauses only.
    patterns = [
        (r"(?:网页)?正文(?:中)?(?:包含|出现)\s*(.+?)(?:\s*错误标记)?", "body", "="),
        (r"标题(?:完整(?:为|等于)|精确(?:为|等于))\s*(.+?)", "title", "=="),
        (r"标题(?:包含|有)\s*(.+?)", "title", "="),
        (r"标题不包含\s*(.+?)", "title", "!="),
        (r"(?:header|响应头)(?:中)?(?:包含|含有|含)\s*(.+?)", "header", "="),
        (r"(?:Banner|服务响应原文|响应原文)(?:中)?(?:包含|含有|含)\s*(.+?)", "banner", "="),
        (r"响应信息(?:中)?(?:出现|有)\s*(.+?)", None, "="),
        (r"引用\s*(.+?)\s*文件", "js_name", "="),
    ]
    for pattern, field, op in patterns:
        m = re.fullmatch(pattern + r"\s*" + END, s, re.I)
        if m and literal_ok(m[1]):
            value = m[1].strip()
            if field is None:
                if value.startswith("SSH-"):
                    field = "banner"
                elif value.startswith(("Content-", "Set-Cookie", "Server:")):
                    field = "header"
                else:
                    return None
            return pred(field, value, op)
    return None


def split_top_level(text, separators):
    """Split only outside explicit parentheses and quoted literals."""
    parts, start, depth, quote, i = [], 0, 0, None, 0
    while i < len(text):
        ch = text[i]
        if quote:
            if ch == "\\":
                i += 2
                continue
            if ch == quote:
                quote = None
        elif ch == '"':
            quote = ch
        elif ch == "(":
            depth += 1
        elif ch == ")":
            depth -= 1
            if depth < 0:
                return None
        elif depth == 0:
            found = next((token for token in separators if text.startswith(token, i)), None)
            if found:
                parts.append(text[start:i].strip())
                i += len(found)
                start = i
                continue
        i += 1
    if depth != 0 or quote:
        return None
    parts.append(text[start:].strip())
    return parts


def expression(text, depth=0):
    if depth > 16:
        return unresolved(text, "too_complex")
    direct = atom(text)
    if direct:
        return direct
    if "最安全" in text:
        return unresolved(text, "ambiguous_intent")
    if "容器" in text and "内存" in text and re.search(r"没有.*(?:监控|权限)", text):
        return unresolved(text, "unobservable_intent")
    if re.search(r"转换|转成|修正|https?://|CVE-", text, re.I):
        return unresolved(text)
    s = text.replace("（", "(").replace("）", ")")
    s = re.sub(r"^同一条资产记录[，,]\s*", "", s)
    # OR has lower precedence; mixed scope should use explicit parentheses.
    for separators, op in [(("或者", "或"), "||"), (("，并且", "，而且", "，同时", "，且", ",并且", ",同时", ",且", "并且", "而且", "同时", "且", "，", ","), "&&")]:
        parts = split_top_level(s, separators)
        if parts is None:
            return unresolved(text, "invalid_structure")
        if len(parts) > 1:
            return group(op, [expression(p, depth + 1) for p in parts])
    wrapped = re.fullmatch(r"\((.*)\)" + END, s, re.S)
    if wrapped:
        return expression(wrapped[1].strip(), depth + 1)
    return unresolved(text)


def parse_conditions(text):
    """Always return JSON-serializable conditions, including unresolved intent."""
    if not isinstance(text, str) or not text.strip() or len(text) > 20000:
        node = unresolved("", "invalid_input")
    else:
        node = expression(clean(text))
    return {"version": "1.0", "condition": node}

