package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Silas-Go/fofa-query-agent/internal/agent"
	"github.com/Silas-Go/fofa-query-agent/internal/contest"
)

func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
}

func NewHandler(page, examples []byte, templates ...[]byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		method := http.MethodGet
		switch r.URL.Path {
		case "/", "/api/examples", "/health":
		case "/api/answer", "/api/submission":
			method = http.MethodPost
		default:
			reply(w, 404, map[string]string{"error": "接口不存在"})
			return
		}
		if r.Method != method {
			w.Header().Set("Allow", method)
			reply(w, 405, map[string]string{"error": "请求方法不支持"})
			return
		}
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(page)
		case "/api/examples":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, _ = w.Write(examples)
		case "/health":
			reply(w, 200, map[string]string{"status": "ok", "version": agent.Version})
		case "/api/answer", "/api/submission":
			if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host {
				reply(w, 403, map[string]string{"error": "只允许本地页面请求"})
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 2_000_000)
			defer r.Body.Close()
			data, err := io.ReadAll(r.Body)
			if err != nil {
				reply(w, 413, map[string]string{"error": "请求超过2 MB或读取失败"})
				return
			}
			var questions []agent.Question
			if err = agent.DecodeJSON(data, &questions); err != nil {
				reply(w, 400, map[string]string{"error": err.Error()})
				return
			}
			if r.URL.Path == "/api/submission" {
				if len(templates) == 0 {
					reply(w, 500, map[string]string{"error": "缺少答案模板"})
					return
				}
				var sheet contest.Submission
				if err = agent.DecodeJSON(templates[0], &sheet); err != nil {
					reply(w, 500, map[string]string{"error": err.Error()})
					return
				}
				sheet, err = contest.Generate(questions, sheet)
				if err != nil {
					reply(w, 400, map[string]string{"error": err.Error()})
					return
				}
				data, err = contest.Encode(sheet)
				if err != nil {
					reply(w, 400, map[string]string{"error": err.Error()})
					return
				}
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				_, _ = w.Write(data)
				return
			}
			answers, err := agent.ProcessBatch(questions)
			if err != nil {
				reply(w, 400, map[string]string{"error": err.Error()})
				return
			}
			reply(w, 200, map[string]any{"results": answers})
		}
	})
}
