// Package agent implements the natural-language -> JSON -> validation -> query pipeline.
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

const Version = "0.4.0"

type Condition struct {
	Type       string      `json:"type"`
	Field      string      `json:"field,omitempty"`
	Operator   string      `json:"operator,omitempty"`
	Value      any         `json:"value"`
	Conditions []Condition `json:"conditions,omitempty"`
	Text       string      `json:"text,omitempty"`
	Reason     string      `json:"reason,omitempty"`
}

type Document struct {
	Version   string     `json:"version"`
	Condition *Condition `json:"condition"`
}

type Issue struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type Validation struct {
	Valid      bool      `json:"valid"`
	Errors     []Issue   `json:"errors"`
	Normalized *Document `json:"normalized"`
}

type Question struct {
	ID   string `json:"题号"`
	Text string `json:"自然语言输入"`
}

type Evidence struct {
	Source string `json:"来源"`
	Reason string `json:"依据"`
	Date   string `json:"核对日期"`
}

type Check struct {
	Valid  bool    `json:"通过"`
	Errors []Issue `json:"错误"`
}

type Answer struct {
	Question
	Status     string     `json:"状态"`
	Query      *string    `json:"查询语句"`
	Note       string     `json:"说明"`
	Evidence   []Evidence `json:"依据"`
	Conditions *Document  `json:"结构化条件"`
	Check      Check      `json:"校验结果"`
}

// DecodeJSON rejects unknown properties, trailing JSON and float coercion.
func DecodeJSON(data []byte, target any) error {
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	if !utf8.Valid(data) {
		return fmt.Errorf("文件必须使用UTF-8编码")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("JSON格式错误：%w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("JSON之后不能包含额外内容")
	}
	return nil
}

func predicate(field string, value any, op string) Condition {
	operator := "eq"
	if op == "=" && fields[field].Kind == "text" {
		operator = "contains"
	} else if op == "!=" {
		operator = "ne"
		if fields[field].Kind == "text" {
			operator = "not_contains"
		}
	}
	return Condition{Type: "predicate", Field: field, Operator: operator, Value: value}
}

func group(kind string, children []Condition) Condition {
	if len(children) == 1 {
		return children[0]
	}
	return Condition{Type: kind, Conditions: children}
}

func unresolved(text, reason string) Condition {
	return Condition{Type: "unresolved", Text: text, Reason: reason}
}
