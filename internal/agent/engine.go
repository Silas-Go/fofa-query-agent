package agent

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func ValidateBatch(items []Question) error {
	if len(items) < 1 || len(items) > 1000 {
		return fmt.Errorf("每批需要1–1000道题")
	}
	seen := map[string]bool{}
	for i, item := range items {
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Text) == "" {
			return fmt.Errorf("第%d条的题号和自然语言输入必须为非空字符串", i+1)
		}
		if utf8.RuneCountInString(item.Text) > 20000 {
			return fmt.Errorf("第%d条题目超过20000字符", i+1)
		}
		if seen[item.ID] {
			return fmt.Errorf("题号重复：%s", item.ID)
		}
		seen[item.ID] = true
	}
	return nil
}

func Process(item Question) Answer {
	document := Parse(item.Text)
	validation := Validate(document)
	result := Answer{Question: item, Status: validation.Status(), Conditions: &document, Evidence: []Evidence{}, Check: Check{Valid: validation.Valid, Errors: validation.Errors}}
	if !validation.Valid {
		parts := []string{}
		for _, issue := range validation.Errors {
			parts = append(parts, issue.Message)
		}
		result.Note = strings.Join(parts, "；")
		return result
	}
	result.Conditions = validation.Normalized
	query, err := Generate(*result.Conditions)
	if err != nil {
		result.Status = "依赖失败"
		result.Check.Valid = false
		result.Note = err.Error()
		return result
	}
	result.Query = &query
	result.Note = "结构化条件已通过本地校验；查询由条件树生成，未执行FOFA搜索。"
	if regex("(?i)B\\s*段").MatchString(item.Text) {
		result.Note += "“B段”按/16网段解释。"
	}
	result.Evidence = []Evidence{{Source: "https://fofa.info/", Reason: "FOFA官方搜索语法表；结构化条件通过Go校验器后生成，未执行资产查询。", Date: "2026-09-30"}}
	return result
}

func ProcessBatch(items []Question) ([]Answer, error) {
	if err := ValidateBatch(items); err != nil {
		return nil, err
	}
	answers := make([]Answer, len(items))
	for i, item := range items {
		answers[i] = Process(item)
	}
	return answers, nil
}
