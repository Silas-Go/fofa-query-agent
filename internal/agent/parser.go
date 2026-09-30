package agent

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
)

const ending = `\s*(?:的)?(?:全部|所有)?(?:资产记录|网站类资产|资产|网站|网页|数据)?`
const domainPattern = `[A-Za-z0-9](?:[A-Za-z0-9.-]*[A-Za-z0-9])?\.[A-Za-z]{2,}`

var regexCache sync.Map

func regex(pattern string) *regexp.Regexp {
	if compiled, ok := regexCache.Load(pattern); ok {
		return compiled.(*regexp.Regexp)
	}
	compiled := regexp.MustCompile(pattern)
	actual, _ := regexCache.LoadOrStore(pattern, compiled)
	return actual.(*regexp.Regexp)
}

func match(pattern, text string) []string {
	return regex(`(?is)^(?:` + pattern + `)$`).FindStringSubmatch(text)
}

func clean(text string) string {
	s := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(text), "。！？?!"))
	return strings.TrimSpace(regex(`^(?:请帮我查询|请帮我查|帮我找一下|我想要搜索|我想要查询|我想搜索|我想查询|请查询|请搜索|请帮我找|帮我查询|我想找|我想查|我想看|帮我找|搜索|查询|查找|找|查)\s*`).ReplaceAllString(s, ""))
}

func country(value string) string {
	s := strings.TrimSpace(value)
	if code := countries[s]; code != "" {
		return code
	}
	if regex(`^[A-Z]{2}$`).MatchString(s) {
		return s
	}
	return ""
}

func literalOK(s string) bool {
	return strings.TrimSpace(s) != "" && !regex(`[，,；;\n]|或者|或|并且|且|同时|但是|但|要求|任意|至少|不包含|不限定`).MatchString(s)
}

type atomRule struct{ pattern, field, op string }

func atom(text string) (Condition, bool) {
	s := strings.TrimRight(strings.TrimSpace(text), "。")
	if m := match(`(?:网页源码|网页正文)中包含这段原样文本(?:的资产)?[：:](.+)`, s); m != nil {
		return predicate("body", m[1], "="), true
	}
	if m := match(`字段\s*([A-Za-z_][\w.]*)\s*(?:为|等于)\s*(.+?)`+ending, s); m != nil && literalOK(m[2]) {
		return Condition{Type: "predicate", Field: m[1], Operator: "eq", Value: m[2]}, true
	}
	if m := match(`(?:IP\s*(?:地址)?\s*(?:为|是)?\s*|地址或网段(?:为|是)\s*)([\d.]+(?:/\d+)?)`+ending, s); m != nil {
		return predicate("ip", m[1], "="), true
	}
	if m := match(`([\d.]+/\d+)\s*网段内`+ending, s); m != nil {
		return predicate("ip", m[1], "="), true
	}
	if m := match(`(?:这个\s*IP\s*的)?整个\s*B\s*段[：:]\s*([\d.]+)`, s); m != nil {
		return predicate("ip", m[1]+"/16", "="), true
	}
	for _, rule := range []atomRule{
		{`主域(?:是|为)\s*(` + domainPattern + `)`, "domain", "="},
		{`网站名(?:是|为)\s*(` + domainPattern + `)`, "host", "=="},
		{`域名(?:里)?(?:带有|包含)\s*(\S+?)`, "host", "="},
		{`自治系统号(?:为|是)\s*(\d+)`, "asn", "="},
		{`操作系统(?:被)?识别为\s*(\S+?)`, "os", "="},
		{`网络组织名称(?:里)?包含\s*(\S+?)`, "org", "="},
	} {
		if m := match(rule.pattern+`\s*`+ending, s); m != nil && literalOK(m[1]) {
			var value any = strings.TrimSpace(m[1])
			if rule.field == "asn" {
				value = json.Number(m[1])
			}
			return predicate(rule.field, value, rule.op), true
		}
	}
	if m := match(`(?:开放\s*(-?\d+)\s*端口|端口(?:为|是)\s*(-?\d+))`+ending, s); m != nil {
		v := m[1]
		if v == "" {
			v = m[2]
		}
		return predicate("port", json.Number(v), "="), true
	}
	if m := match(`(?:网页)?状态码(?:为|是)\s*(\d+)`+ending, s); m != nil {
		n := predicate("status_code", json.Number(m[1]), "=")
		if strings.Contains(s, "网站类资产") {
			n = group("and", []Condition{n, predicate("type", "subdomain", "=")})
		}
		return n, true
	}
	if m := match(`(?:使用\s*)?(TCP|UDP)\s*(?:协议|传输协议)`+ending, s); m != nil {
		return predicate("base_protocol", strings.ToLower(m[1]), "="), true
	}
	if m := match(`(DNS|SSH|FTP|SNMP|Memcached)\s*(?:协议|服务|缓存服务协议)`+ending, s); m != nil {
		return predicate("protocol", strings.ToLower(m[1]), "="), true
	}
	if m := match(`(.+?)的\s*(DNS|SSH|FTP|SNMP|Memcached)\s*服务`+ending, s); m != nil && country(m[1]) != "" {
		return group("and", []Condition{predicate("country", country(m[1]), "="), predicate("protocol", strings.ToLower(m[2]), "=")}), true
	}
	if m := match(`(?:位于\s*|国家(?:代码)?(?:又?必须)?(?:等于|为|是)\s*|地域限定为\s*)(.+?)`+ending, s); m != nil && country(m[1]) != "" {
		return predicate("country", country(m[1]), "="), true
	}
	if m := match(`国家(?:代码)?(?:又?必须)?(?:不等于|不是)\s*(.+?)`+ending, s); m != nil && country(m[1]) != "" {
		return predicate("country", country(m[1]), "!="), true
	}
	if match(`(?:不是|非)\s*IPv6`+ending, s) != nil {
		return predicate("is_ipv6", false, "="), true
	}
	if m := match(`(已|没有|未)绑定域名`+ending, s); m != nil {
		return predicate("is_domain", m[1] == "已", "="), true
	}
	for _, rule := range []atomRule{
		{`(?:网页)?正文(?:中)?(?:包含|出现)\s*(.+?)(?:\s*错误标记)?`, "body", "="},
		{`标题(?:完整(?:为|等于)|精确(?:为|等于))\s*(.+?)`, "title", "=="},
		{`标题(?:包含|有)\s*(.+?)`, "title", "="},
		{`标题不包含\s*(.+?)`, "title", "!="},
		{`(?:header|响应头)(?:中)?(?:包含|含有|含)\s*(.+?)`, "header", "="},
		{`(?:Banner|服务响应原文|响应原文)(?:中)?(?:包含|含有|含)\s*(.+?)`, "banner", "="},
		{`响应信息(?:中)?(?:出现|有)\s*(.+?)`, "", "="},
		{`引用\s*(.+?)\s*文件`, "js_name", "="},
	} {
		if m := match(rule.pattern+`\s*`+ending, s); m != nil && literalOK(m[1]) {
			value := strings.TrimSpace(m[1])
			field := rule.field
			if field == "" {
				switch {
				case strings.HasPrefix(value, "SSH-"):
					field = "banner"
				case strings.HasPrefix(value, "Content-"), strings.HasPrefix(value, "Set-Cookie"), strings.HasPrefix(value, "Server:"):
					field = "header"
				default:
					return Condition{}, false
				}
			}
			return predicate(field, value, rule.op), true
		}
	}
	return Condition{}, false
}

func splitTop(text string, separators []string) ([]string, bool) {
	parts := []string{}
	start, depth := 0, 0
	quoted := false
	for i := 0; i < len(text); {
		ch := text[i]
		if quoted {
			if ch == '\\' {
				i++
				if i < len(text) {
					_, n := utf8.DecodeRuneInString(text[i:])
					i += n
				}
				continue
			}
			if ch == '"' {
				quoted = false
			}
		} else if ch == '"' {
			quoted = true
		} else if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth < 0 {
				return nil, false
			}
		} else if depth == 0 {
			found := ""
			for _, sep := range separators {
				if strings.HasPrefix(text[i:], sep) {
					found = sep
					break
				}
			}
			if found != "" {
				parts = append(parts, strings.TrimSpace(text[start:i]))
				i += len(found)
				start = i
				continue
			}
		}
		_, n := utf8.DecodeRuneInString(text[i:])
		i += n
	}
	if depth != 0 || quoted {
		return nil, false
	}
	return append(parts, strings.TrimSpace(text[start:])), true
}

func expression(text string, depth int) Condition {
	if depth > 16 {
		return unresolved(text, "too_complex")
	}
	if n, ok := atom(text); ok {
		return n
	}
	if strings.Contains(text, "最安全") {
		return unresolved(text, "ambiguous_intent")
	}
	if strings.Contains(text, "容器") && strings.Contains(text, "内存") && regex(`没有.*(?:监控|权限)`).MatchString(text) {
		return unresolved(text, "unobservable_intent")
	}
	if regex(`(?i)转换|转成|修正|https?://|CVE-`).MatchString(text) {
		return unresolved(text, "unsupported_intent")
	}
	s := strings.NewReplacer("（", "(", "）", ")").Replace(text)
	s = regex(`^同一条资产记录[，,]\s*`).ReplaceAllString(s, "")
	for i, seps := range [][]string{{"或者", "或"}, {"，并且", "，而且", "，同时", "，且", ",并且", ",同时", ",且", "并且", "而且", "同时", "且", "，", ","}} {
		parts, ok := splitTop(s, seps)
		if !ok {
			return unresolved(text, "invalid_structure")
		}
		if len(parts) > 1 {
			children := make([]Condition, 0, len(parts))
			for _, p := range parts {
				children = append(children, expression(p, depth+1))
			}
			kind := "and"
			if i == 0 {
				kind = "or"
			}
			return group(kind, children)
		}
	}
	if m := match(`\((.*)\)`+ending, s); m != nil {
		return expression(strings.TrimSpace(m[1]), depth+1)
	}
	return unresolved(text, "unsupported_intent")
}

// Parse generates structured data only; invalid values remain for the validator.
func Parse(text string) Document {
	if document, ok := reviewedDocument(text); ok {
		return document
	}
	n := unresolved("", "invalid_input")
	if strings.TrimSpace(text) != "" && utf8.RuneCountInString(text) <= 20000 {
		n = expression(clean(text), 0)
	}
	return Document{Version: "1.0", Condition: &n}
}
