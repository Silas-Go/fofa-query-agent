package agent

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/netip"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

var intentErrors = map[string]Issue{
	"unsupported_intent":  {Code: "UNSUPPORTED_INTENT", Status: "暂未支持", Message: "当前解析器无法完整表达这段意图，不能省略它继续生成。"},
	"ambiguous_intent":    {Code: "AMBIGUOUS_INTENT", Status: "需要澄清", Message: "请给出“最安全”等描述的具体可检索标准。"},
	"unobservable_intent": {Code: "UNEXPRESSIBLE_INTENT", Status: "能力不支持", Message: "缺少容器监控数据或运行时访问权限，资产搜索无法确定实时内存使用率。"},
	"invalid_structure":   {Code: "INVALID_STRUCTURE", Status: "输入非法", Message: "括号或引号不完整，无法确定条件分组。"},
	"invalid_input":       {Code: "INVALID_INPUT", Status: "输入非法", Message: "自然语言输入必须为1–20000字符的非空文本。"},
	"too_complex":         {Code: "UNSUPPORTED_COMPLEXITY", Status: "暂未支持", Message: "条件嵌套超过当前处理上限。"},
}

type checker struct {
	issues []Issue
	count  int
}

func (c *checker) add(code, message, path, status string) {
	c.issues = append(c.issues, Issue{Code: code, Message: message, Path: path, Status: status})
}

func (c *checker) visit(n Condition, path string, depth int) Condition {
	c.count++
	if c.count > 256 || depth > 16 {
		c.add("UNSUPPORTED_COMPLEXITY", "最多支持256个节点、16层嵌套。", path, "暂未支持")
		return n
	}
	switch n.Type {
	case "unresolved":
		issue, ok := intentErrors[n.Reason]
		if !ok || n.Field != "" || n.Operator != "" || n.Value != nil || n.Conditions != nil {
			c.add("INVALID_STRUCTURE", "未解析意图节点格式错误。", path, "输入非法")
		} else {
			issue.Path = path
			c.issues = append(c.issues, issue)
		}
		return n
	case "and", "or":
		if len(n.Conditions) < 2 || len(n.Conditions) > 64 || n.Field != "" || n.Operator != "" || n.Value != nil || n.Text != "" || n.Reason != "" {
			c.add("INVALID_STRUCTURE", "组合节点必须包含2–64个子条件，且不能混入叶子属性。", path, "输入非法")
			return n
		}
		children := make([]Condition, len(n.Conditions))
		for i, child := range n.Conditions {
			children[i] = c.visit(child, fmt.Sprintf("%s.conditions[%d]", path, i), depth+1)
		}
		n.Conditions = children
		return n
	case "predicate":
		if n.Conditions != nil || n.Text != "" || n.Reason != "" {
			c.add("INVALID_STRUCTURE", "叶子条件不能携带组合或未解析属性。", path, "输入非法")
			return n
		}
	default:
		c.add("INVALID_STRUCTURE", "只接受predicate、and、or和unresolved节点。", path, "输入非法")
		return n
	}
	rule, ok := fields[n.Field]
	if !ok {
		c.add("UNSUPPORTED_FIELD", "当前字段白名单不支持："+n.Field, path+".field", "暂未支持")
		return n
	}
	if _, ok := rule.Operators[n.Operator]; !ok {
		c.add("UNSUPPORTED_OPERATOR", "字段不支持该操作："+n.Field+" / "+n.Operator, path+".operator", "暂未支持")
		return n
	}
	if n.Operator == "empty" || n.Operator == "not_empty" {
		if value, ok := n.Value.(string); !ok || value != "" {
			c.add("INVALID_VALUE", "空值操作必须使用空字符串。", path+".value", "输入非法")
		}
		return n
	}
	if n.Operator == "wildcard" && rule.Kind == "domain" {
		rule.Kind = "text"
	}
	value, err := normalizeValue(n.Value, rule)
	if err != nil {
		c.add("INVALID_VALUE", n.Field+"："+err.Error(), path+".value", "输入非法")
		return n
	}
	n.Value = value
	return n
}

func normalizeValue(value any, rule fieldRule) (any, error) {
	if rule.Kind == "integer" {
		var number int64
		switch v := value.(type) {
		case int:
			number = int64(v)
		case int64:
			number = v
		case json.Number:
			parsed, err := v.Int64()
			if err != nil {
				return nil, fmt.Errorf("必须为整数")
			}
			number = parsed
		default:
			return nil, fmt.Errorf("必须为整数")
		}
		if number < rule.Min || number > rule.Max {
			return nil, fmt.Errorf("必须在%d–%d范围内", rule.Min, rule.Max)
		}
		return number, nil
	}
	if rule.Kind == "boolean" {
		if b, ok := value.(bool); ok {
			return b, nil
		}
		return nil, fmt.Errorf("必须为true或false")
	}
	s, ok := value.(string)
	if !ok || (s == "" && !rule.AllowEmpty) || utf8.RuneCountInString(s) > 20000 {
		return nil, fmt.Errorf("必须是非空字符串，且不超过20000字符")
	}
	switch rule.Kind {
	case "date":
		if _, err := time.Parse("2006-01-02", s); err != nil {
			if _, err = time.Parse("2006-01-02 15:04:05", s); err != nil {
				return nil, fmt.Errorf("日期格式错误")
			}
		}
	case "ip":
		if strings.Contains(s, "/") {
			p, err := netip.ParsePrefix(s)
			if err != nil {
				return nil, fmt.Errorf("不是合法IP或CIDR")
			}
			return p.Masked().String(), nil
		}
		a, err := netip.ParseAddr(s)
		if err != nil || a.Zone() != "" {
			return nil, fmt.Errorf("不是合法IP地址")
		}
		return a.String(), nil
	case "domain":
		if len(s) > 253 || !regex("^"+domainPattern+"$").MatchString(s) {
			return nil, fmt.Errorf("不是有效域名格式")
		}
		for _, label := range strings.Split(s, ".") {
			if !regex("^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$").MatchString(label) {
				return nil, fmt.Errorf("不是有效域名标签")
			}
		}
		return strings.ToLower(s), nil
	case "enum":
		if !rule.Values[s] {
			return nil, fmt.Errorf("值不在当前支持的枚举范围：%s", s)
		}
	}
	return s, nil
}

// Validate checks the entire tree before any rendering can take place.
func Validate(document Document) Validation {
	c := checker{issues: []Issue{}}
	if document.Version != "1.0" || document.Condition == nil {
		c.add("INVALID_STRUCTURE", "需要version=1.0和非空condition。", "$", "输入非法")
		return Validation{Errors: c.issues}
	}
	root := c.visit(*document.Condition, "$.condition", 0)
	if len(c.issues) == 0 {
		options, err := branches(root)
		if err != nil {
			c.add("UNSUPPORTED_COMPLEXITY", err.Error(), "$.condition", "暂未支持")
		} else {
			allConflict := true
			for _, branch := range options {
				if !branchConflict(branch) {
					allConflict = false
					break
				}
			}
			if allConflict {
				c.add("CONFLICT", "所有可选分支都存在无法同时满足的条件。", "$.condition", "条件矛盾")
			}
		}
	}
	result := Validation{Valid: len(c.issues) == 0, Errors: c.issues}
	if result.Valid {
		result.Normalized = &Document{Version: "1.0", Condition: &root}
	}
	return result
}

func branches(n Condition) ([][]Condition, error) {
	if n.Type == "predicate" {
		return [][]Condition{{n}}, nil
	}
	result := [][]Condition{}
	if n.Type == "and" {
		result = append(result, []Condition{})
	}
	for _, child := range n.Conditions {
		options, err := branches(child)
		if err != nil {
			return nil, err
		}
		if n.Type == "or" {
			if len(result)+len(options) > 128 {
				return nil, fmt.Errorf("组合展开超过128个分支")
			}
			result = append(result, options...)
		} else {
			if len(result)*len(options) > 128 {
				return nil, fmt.Errorf("组合展开超过128个分支")
			}
			next := [][]Condition{}
			for _, a := range result {
				for _, b := range options {
					joined := append([]Condition{}, a...)
					joined = append(joined, b...)
					next = append(next, joined)
				}
			}
			result = next
		}
	}
	return result, nil
}

func branchConflict(leaves []Condition) bool {
	byField := map[string][]Condition{}
	for _, p := range leaves {
		byField[p.Field] = append(byField[p.Field], p)
	}
	for name, predicates := range byField {
		if name == "ip" {
			continue
		}
		eq, ne, required, forbidden := map[any]bool{}, map[any]bool{}, map[any]bool{}, map[any]bool{}
		for _, p := range predicates {
			switch p.Operator {
			case "eq":
				eq[p.Value] = true
			case "ne":
				ne[p.Value] = true
			case "contains":
				required[p.Value] = true
			case "not_contains":
				forbidden[p.Value] = true
			case "empty":
				eq[""] = true
			case "not_empty":
				ne[""] = true
			}
		}
		if len(eq) > 1 {
			return true
		}
		for v := range eq {
			if ne[v] || forbidden[v] {
				return true
			}
		}
		for v := range required {
			if forbidden[v] {
				return true
			}
		}
	}
	return ipConflict(byField)
}

type interval struct {
	low, high *big.Int
	bits      int
}

func ipRange(value string) interval {
	var p netip.Prefix
	if strings.Contains(value, "/") {
		p, _ = netip.ParsePrefix(value)
	} else {
		a, _ := netip.ParseAddr(value)
		p = netip.PrefixFrom(a, a.BitLen())
	}
	p = p.Masked()
	low := new(big.Int).SetBytes(p.Addr().AsSlice())
	size := new(big.Int).Lsh(big.NewInt(1), uint(p.Addr().BitLen()-p.Bits()))
	high := new(big.Int).Sub(new(big.Int).Add(low, size), big.NewInt(1))
	return interval{low: low, high: high, bits: p.Addr().BitLen()}
}

func ipConflict(byField map[string][]Condition) bool {
	if len(byField["ip"]) == 0 {
		return false
	}
	for _, bits := range []int{32, 128} {
		low := big.NewInt(0)
		high := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
		possible := true
		excluded := []interval{}
		for _, flag := range byField["is_ipv6"] {
			if flag.Value.(bool) != (bits == 128) {
				possible = false
			}
		}
		for _, p := range byField["ip"] {
			r := ipRange(p.Value.(string))
			if p.Operator == "eq" {
				if r.bits != bits {
					possible = false
					continue
				}
				if r.low.Cmp(low) > 0 {
					low.Set(r.low)
				}
				if r.high.Cmp(high) < 0 {
					high.Set(r.high)
				}
			} else if r.bits == bits {
				excluded = append(excluded, r)
			}
		}
		if !possible || low.Cmp(high) > 0 {
			continue
		}
		sort.Slice(excluded, func(i, j int) bool { return excluded[i].low.Cmp(excluded[j].low) < 0 })
		cursor := new(big.Int).Set(low)
		for _, r := range excluded {
			if r.high.Cmp(cursor) < 0 {
				continue
			}
			if r.low.Cmp(cursor) > 0 {
				break
			}
			cursor.Add(r.high, big.NewInt(1))
		}
		if cursor.Cmp(high) <= 0 {
			return false
		}
	}
	return true
}

func (v Validation) Status() string {
	if v.Valid {
		return "成功"
	}
	for _, status := range []string{"输入非法", "条件矛盾", "需要澄清", "能力不支持", "暂未支持"} {
		for _, issue := range v.Errors {
			if issue.Status == status {
				return status
			}
		}
	}
	return "依赖失败"
}
