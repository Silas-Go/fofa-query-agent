package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Silas-Go/fofa-query-agent/internal/agent"
	"github.com/Silas-Go/fofa-query-agent/internal/server"
)

//go:embed web/index.html
var page []byte

//go:embed fixtures/questions.json
var examples []byte

func output(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(value)
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法：fofa-query-agent serve | query 文本 | parse 文本 | compile 文件 | answer 文件 [-o 输出文件]")
	}
	switch args[0] {
	case "query", "parse":
		if len(args) < 2 {
			return fmt.Errorf("请提供自然语言输入")
		}
		text := strings.Join(args[1:], " ")
		if args[0] == "parse" {
			return output(os.Stdout, agent.Parse(text))
		}
		return output(os.Stdout, agent.Process(agent.Question{ID: "CLI-001", Text: text}))
	case "compile":
		if len(args) != 2 {
			return fmt.Errorf("请提供一个条件JSON文件")
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		var document agent.Document
		if err = agent.DecodeJSON(data, &document); err != nil {
			return err
		}
		query, err := agent.Generate(document)
		if err != nil {
			return err
		}
		fmt.Println(query)
		return nil
	case "answer":
		if len(args) != 2 && (len(args) != 4 || args[2] != "-o") {
			return fmt.Errorf("用法：answer 输入文件 [-o 输出文件]")
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		var questions []agent.Question
		if err = agent.DecodeJSON(data, &questions); err != nil {
			return err
		}
		answers, err := agent.ProcessBatch(questions)
		if err != nil {
			return err
		}
		if len(args) == 2 {
			return output(os.Stdout, answers)
		}
		if err = os.MkdirAll(filepath.Dir(args[3]), 0755); err != nil {
			return err
		}
		file, err := os.Create(args[3])
		if err != nil {
			return err
		}
		writeErr := output(file, answers)
		closeErr := file.Close()
		if writeErr != nil {
			return writeErr
		}
		return closeErr
	case "serve":
		flags := flag.NewFlagSet("serve", flag.ContinueOnError)
		port := flags.Int("port", 8765, "本地服务端口")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *port < 1 || *port > 65535 {
			return fmt.Errorf("请指定1–65535范围内的端口")
		}
		srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", *port), Handler: server.NewHandler(page, examples), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		go func() {
			<-ctx.Done()
			shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = srv.Shutdown(shutdown)
		}()
		fmt.Printf("FOFA 查询助手：http://%s\n按 Ctrl+C 停止。\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	default:
		return fmt.Errorf("未知命令：%s", args[0])
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
