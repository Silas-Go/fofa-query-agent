package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Silas-Go/fofa-query-agent/internal/agent"
)

func TestAnswerEndpoint(t *testing.T) {
	handler := NewHandler([]byte("FOFA"), []byte("[]"))
	request := httptest.NewRequest("POST", "/api/answer", strings.NewReader(`[{"题号":"1","自然语言输入":"端口为443的资产"}]`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	var body struct {
		Results []agent.Answer `json:"results"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || len(body.Results) != 1 || body.Results[0].Query == nil || *body.Results[0].Query != `port="443"` {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}
