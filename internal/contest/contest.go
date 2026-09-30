// Package contest validates and builds the exact competition upload format.
package contest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/Silas-Go/fofa-query-agent/internal/agent"
)

const Refusal = "该需求不能直接转换为FOFA搜索语句"
const MaxBytes = 8 * 1024 * 1024

type Entry struct {
	ID    string `json:"题号"`
	Query string `json:"查询语句"`
}

type Submission struct {
	Player    string  `json:"选手名称"`
	PackageID string  `json:"参赛包编号"`
	Answers   []Entry `json:"答案"`
}

func Check(questions []agent.Question, sheet Submission, allowBlank bool) error {
	if err := agent.ValidateBatch(questions); err != nil {
		return err
	}
	if strings.TrimSpace(sheet.Player) == "" {
		return fmt.Errorf("选手名称不能为空")
	}
	if !strings.HasPrefix(sheet.PackageID, "pkg-") || len(sheet.PackageID) <= 4 {
		return fmt.Errorf("参赛包编号缺失或格式错误")
	}
	if len(sheet.Answers) != len(questions) {
		return fmt.Errorf("答卷%d题，题目%d题：数量不一致", len(sheet.Answers), len(questions))
	}
	expected := map[string]bool{}
	for _, q := range questions {
		expected[q.ID] = true
	}
	seen := map[string]bool{}
	for _, entry := range sheet.Answers {
		if !expected[entry.ID] {
			return fmt.Errorf("未知题号：%s", entry.ID)
		}
		if seen[entry.ID] {
			return fmt.Errorf("题号重复：%s", entry.ID)
		}
		seen[entry.ID] = true
		if !allowBlank && strings.TrimSpace(entry.Query) == "" {
			return fmt.Errorf("题号%s的查询为空", entry.ID)
		}
	}
	return nil
}

// Generate does not turn implementation gaps or runtime failures into refusals.
func Generate(questions []agent.Question, template Submission) (Submission, error) {
	if err := Check(questions, template, true); err != nil {
		return Submission{}, err
	}
	results, err := agent.ProcessBatch(questions)
	if err != nil {
		return Submission{}, err
	}
	queries := map[string]string{}
	for _, result := range results {
		if result.Query != nil && result.Check.Valid {
			queries[result.ID] = *result.Query
			continue
		}
		if len(result.Check.Errors) == 0 {
			return Submission{}, fmt.Errorf("%s：没有查询，也没有明确拒绝原因", result.ID)
		}
		for _, issue := range result.Check.Errors {
			switch issue.Code {
			case "INVALID_VALUE", "CONFLICT", "AMBIGUOUS_INTENT", "UNEXPRESSIBLE_INTENT":
			default:
				return Submission{}, fmt.Errorf("%s尚未完成：%s；不能当成无法转换提交", result.ID, issue.Message)
			}
		}
		queries[result.ID] = Refusal
	}
	output := Submission{Player: template.Player, PackageID: template.PackageID, Answers: make([]Entry, len(template.Answers))}
	for i, entry := range template.Answers {
		output.Answers[i] = Entry{ID: entry.ID, Query: queries[entry.ID]}
	}
	return output, Check(questions, output, false)
}

func Encode(sheet Submission) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(sheet); err != nil {
		return nil, err
	}
	if buffer.Len() > MaxBytes {
		return nil, fmt.Errorf("答案超过8 MB")
	}
	if !utf8.Valid(buffer.Bytes()) {
		return nil, fmt.Errorf("答案必须为UTF-8")
	}
	return buffer.Bytes(), nil
}

// Save atomically replaces a completed answer file; errors never leave half JSON.
func Save(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".answer-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
