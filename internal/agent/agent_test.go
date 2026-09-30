package agent

import "testing"

func TestCorePipeline(t *testing.T) {
	cases := []struct{ name, text, status, query string }{
		{"normal", "请查询 IP 地址为 20.247.40.92 的资产。", "成功", "ip=\"20.247.40.92\""},
		{"and", "搜索 SSH 协议且端口为 22022 的资产。", "成功", "(protocol=\"ssh\" && port=\"22022\")"},
		{"or", "查询（国家为美国或国家为中国）且端口为443的资产。", "成功", "((country=\"US\" || country=\"CN\") && port=\"443\")"},
		{"invalid port", "端口为1000000的资产", "输入非法", ""},
		{"invalid IP", "IP地址为999.1.1.1的资产", "输入非法", ""},
		{"conflict", "国家为美国且国家不等于美国的资产", "条件矛盾", ""},
		{"unknown field", "字段secret为abc", "暂未支持", ""},
		{"unobservable", "查询容器实时内存使用率，但是没有监控和权限", "能力不支持", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Process(Question{ID: tc.name, Text: tc.text})
			if result.Status != tc.status {
				t.Fatalf("status=%s, want %s: %s", result.Status, tc.status, result.Note)
			}
			if tc.query == "" {
				if result.Query != nil || result.Check.Valid {
					t.Fatal("rejected input generated query")
				}
				return
			}
			if result.Query == nil || *result.Query != tc.query {
				t.Fatalf("query=%v, want %s", result.Query, tc.query)
			}
		})
	}
}

func TestGeneratorEnforcesValidation(t *testing.T) {
	node := Condition{Type: "predicate", Field: "secret", Operator: "eq", Value: "abc"}
	if _, err := Generate(Document{Version: "1.0", Condition: &node}); err == nil {
		t.Fatal("unsupported field accepted")
	}
	// Overlapping ranges can match the same asset and must remain satisfiable.
	root := group("and", []Condition{predicate("ip", "10.0.0.0/8", "="), predicate("ip", "10.1.2.3", "=")})
	if _, err := Generate(Document{Version: "1.0", Condition: &root}); err != nil {
		t.Fatal(err)
	}
}
