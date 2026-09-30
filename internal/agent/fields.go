package agent

type fieldRule struct {
	Kind       string
	Operators  map[string]string
	Min, Max   int64
	Values     map[string]bool
	AllowEmpty bool
}

func values(items ...string) map[string]bool {
	out := make(map[string]bool, len(items))
	for _, item := range items {
		out[item] = true
	}
	return out
}

var countries = map[string]string{
	"中国": "CN", "美国": "US", "英国": "GB", "大不列颠及北爱尔兰联合王国": "GB",
	"德国": "DE", "日本": "JP", "韩国": "KR", "荷兰": "NL", "加拿大": "CA", "土耳其": "TR", "新西兰": "NZ",
}

var fields = makeFields()

func makeFields() map[string]fieldRule {
	equality := func() map[string]string { return map[string]string{"eq": "=", "ne": "!="} }
	f := map[string]fieldRule{
		"ip":            {Kind: "ip", Operators: equality()},
		"port":          {Kind: "integer", Operators: equality(), Min: 0, Max: 65535},
		"asn":           {Kind: "integer", Operators: equality(), Min: 0, Max: 4294967295},
		"status_code":   {Kind: "integer", Operators: equality(), Min: 100, Max: 599},
		"domain":        {Kind: "domain", Operators: equality()},
		"country":       {Kind: "enum", Operators: equality(), Values: values("CN", "US", "GB", "DE", "JP", "KR", "NL", "CA", "TR", "NZ")},
		"protocol":      {Kind: "enum", Operators: equality(), Values: values("dns", "ssh", "ftp", "snmp", "memcached")},
		"base_protocol": {Kind: "enum", Operators: equality(), Values: values("tcp", "udp")},
		"type":          {Kind: "enum", Operators: map[string]string{"eq": "="}, Values: values("service", "subdomain")},
		"is_domain":     {Kind: "boolean", Operators: map[string]string{"eq": "="}},
		"is_ipv6":       {Kind: "boolean", Operators: map[string]string{"eq": "="}},
	}
	for _, name := range []string{"host", "os", "org", "title", "body", "header", "banner", "js_name"} {
		ops := map[string]string{"contains": "=", "not_contains": "!="}
		if name == "host" || name == "title" {
			ops["eq"] = "=="
		}
		f[name] = fieldRule{Kind: "text", Operators: ops}
	}
	for _, name := range []string{"region", "city", "server", "app", "product", "product.version", "category", "cloud_name", "cert", "cert.subject.cn", "cert.subject.org", "cert.issuer.cn", "cert.issuer.org", "cert.domain", "cname_domain", "tls.version", "tls.ja3s", "jarm", "cert.sn", "fid", "icon_hash", "header_hash", "body_hash"} {
		f[name] = fieldRule{Kind: "text", Operators: map[string]string{"contains": "=", "not_contains": "!=", "eq": "=="}}
	}
	f["app"] = fieldRule{Kind: "text", Operators: map[string]string{"contains": "="}}
	for _, name := range []string{"host", "domain", "title", "server", "cloud_name", "cert.subject.cn", "cert.domain"} {
		rule := f[name]
		rule.Operators["wildcard"] = "*="
		f[name] = rule
	}
	for _, name := range []string{"title", "banner", "fid", "icon_hash", "header_hash", "body_hash"} {
		rule := f[name]
		rule.AllowEmpty = true
		rule.Operators["empty"] = "=="
		rule.Operators["not_empty"] = "!="
		f[name] = rule
	}
	for _, name := range []string{"is_cloud", "is_fraud", "is_honeypot", "cert.is_equal", "cert.is_valid", "cert.is_match", "cert.is_expired"} {
		f[name] = fieldRule{Kind: "boolean", Operators: map[string]string{"eq": "="}}
	}
	for _, name := range []string{"after", "before", "cert.not_after.after", "cert.not_after.before"} {
		f[name] = fieldRule{Kind: "date", Operators: map[string]string{"eq": "="}}
	}
	rule := f["country"]
	rule.Values["HK"] = true
	rule.Values["MO"] = true
	rule.Values["TW"] = true
	f["country"] = rule
	rule = f["protocol"]
	rule.Values["socks5"] = true
	f["protocol"] = rule
	return f
}
