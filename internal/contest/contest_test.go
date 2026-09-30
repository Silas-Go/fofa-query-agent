package contest

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Silas-Go/fofa-query-agent/internal/agent"
)

func loadPackage(t *testing.T) ([]agent.Question, Submission) {
	t.Helper()
	q, e := os.ReadFile("../../fixtures/questions.json")
	if e != nil {
		t.Fatal(e)
	}
	a, e := os.ReadFile("../../fixtures/answer-template.json")
	if e != nil {
		t.Fatal(e)
	}
	var questions []agent.Question
	var template Submission
	if e = agent.DecodeJSON(q, &questions); e != nil {
		t.Fatal(e)
	}
	if e = agent.DecodeJSON(a, &template); e != nil {
		t.Fatal(e)
	}
	return questions, template
}

func TestPackageSubmission(t *testing.T) {
	questions, template := loadPackage(t)
	sheet, err := Generate(questions, template)
	if err != nil {
		t.Fatal(err)
	}
	data, err := Encode(sheet)
	if err != nil {
		t.Fatal(err)
	}
	if len(sheet.Answers) != 100 || sheet.PackageID != "pkg-08e82c8d" || sheet.Player != "张致远" {
		t.Fatal("lost metadata")
	}
	var wire map[string]json.RawMessage
	if err = json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	if len(wire) != 3 {
		t.Fatal("unexpected top-level fields")
	}
	var rows []map[string]any
	if err = json.Unmarshal(wire["答案"], &rows); err != nil {
		t.Fatal(err)
	}
	for i, row := range rows {
		if len(row) != 2 {
			t.Fatalf("extra field at %d", i)
		}
		if i >= 90 && i <= 94 {
			if row["查询语句"] != Refusal {
				t.Fatalf("missing refusal at %d", i)
			}
		} else if row["查询语句"] == Refusal {
			t.Fatalf("implementation gap disguised as refusal at %d", i)
		}
	}
	// Serialization must preserve empty strings, false booleans and large serials.
	for _, q := range questions {
		d := agent.Parse(q.Text)
		b, e := json.Marshal(d)
		if e != nil {
			t.Fatal(e)
		}
		var restored agent.Document
		if e = agent.DecodeJSON(b, &restored); e != nil {
			t.Fatal(e)
		}
		if agent.Validate(d).Valid != agent.Validate(restored).Valid {
			t.Fatalf("round-trip changed %s", q.ID)
		}
	}
}

func TestSubmissionRejectsIncompleteOrUnknown(t *testing.T) {
	questions, template := loadPackage(t)
	cases := []struct {
		name string
		edit func(*Submission)
	}{
		{"missing", func(s *Submission) { s.Answers = s.Answers[:99] }},
		{"duplicate", func(s *Submission) { s.Answers[1].ID = s.Answers[0].ID }},
		{"unknown", func(s *Submission) { s.Answers[1].ID = "M999-S999" }},
		{"no package", func(s *Submission) { s.PackageID = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := template
			s.Answers = append([]Entry(nil), template.Answers...)
			tc.edit(&s)
			if _, err := Generate(questions, s); err == nil {
				t.Fatal("bad template accepted")
			}
		})
	}
	questions[0].Text = "完全没有实现的随机需求"
	if _, err := Generate(questions, template); err == nil {
		t.Fatal("unimplemented intent exported as refusal")
	}
}
