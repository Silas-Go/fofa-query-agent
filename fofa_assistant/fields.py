"""MVP field allowlist shared by validation and serialization.

Only this registry controls emitted syntax. The larger T2 catalog is research
material; being present there does not automatically enable a field here.
"""

TEXT_FIELDS = {"host", "os", "org", "title", "body", "header", "banner", "js_name"}
COUNTRY_CODES = {"CN", "US", "GB", "DE", "JP", "KR", "NL", "CA", "TR", "NZ"}


def field(kind, operators, **options):
    return {"kind": kind, "operators": operators, **options}


FIELDS = {
    "ip": field("ip", {"eq": "=", "ne": "!="}),
    "port": field("integer", {"eq": "=", "ne": "!="}, minimum=0, maximum=65535),
    "asn": field("integer", {"eq": "=", "ne": "!="}, minimum=0, maximum=4294967295),
    "status_code": field("integer", {"eq": "=", "ne": "!="}, minimum=100, maximum=599),
    "country": field("enum", {"eq": "=", "ne": "!="}, values=COUNTRY_CODES),
    "domain": field("domain", {"eq": "=", "ne": "!="}),
    "protocol": field("enum", {"eq": "=", "ne": "!="}, values={"dns", "ssh", "ftp", "snmp", "memcached"}),
    "base_protocol": field("enum", {"eq": "=", "ne": "!="}, values={"tcp", "udp"}),
    "type": field("enum", {"eq": "="}, values={"service", "subdomain"}),
    "is_domain": field("boolean", {"eq": "="}),
    "is_ipv6": field("boolean", {"eq": "="}),
    **{name: field("text", {"contains": "=", "not_contains": "!="}) for name in TEXT_FIELDS},
}
for name in ("host", "title"):
    FIELDS[name]["operators"]["eq"] = "=="
