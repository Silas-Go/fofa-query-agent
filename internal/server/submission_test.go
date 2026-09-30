package server

import (
	"bytes"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Silas-Go/fofa-query-agent/internal/agent"
	"github.com/Silas-Go/fofa-query-agent/internal/contest"
)

func TestCompetitionExport(t *testing.T) {
	q, err := os.ReadFile("../../fixtures/questions.json")
	if err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile("../../fixtures/answer-template.json")
	if err != nil {
		t.Fatal(err)
	}
	h := NewHandler([]byte("FOFA"), q, template)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest("POST", "/api/submission", bytes.NewReader(q)))
	var sheet contest.Submission
	if response.Code != 200 {
		t.Fatal(response.Body.String())
	}
	if err = agent.DecodeJSON(response.Body.Bytes(), &sheet); err != nil {
		t.Fatal(err)
	}
	if sheet.PackageID != "pkg-08e82c8d" || len(sheet.Answers) != 100 {
		t.Fatal("incomplete export")
	}
	response = httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest("POST", "/api/submission", bytes.NewBufferString(`[{"题号":"Q1","自然语言输入":"端口为443"}]`)))
	if response.Code != 400 {
		t.Fatal("partial export accepted")
	}
}
