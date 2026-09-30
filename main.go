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
	"strings"
	"syscall"
	"time"

	"github.com/Silas-Go/fofa-query-agent/internal/agent"
	"github.com/Silas-Go/fofa-query-agent/internal/contest"
	"github.com/Silas-Go/fofa-query-agent/internal/server"
)

//go:embed web/index.html
var page []byte

//go:embed fixtures/questions.json
var examples []byte

//go:embed fixtures/answer-template.json
var templateData []byte

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
	case "answer", "check":
		if len(args) < 2 {
			return fmt.Errorf("用法：answer 题目.json [-o 答案.json] [--template 模板.json]；check 答案.json [--questions 题目.json]")
		}
		flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
		out := flags.String("o", "答案.json", "输出文件")
		templatePath := flags.String("template", "", "答案模板；默认使用本参赛包")
		questionPath := flags.String("questions", "", "检查时使用的原题；默认本包")
		name := flags.String("name", "", "覆盖选手名称")
		if err := flags.Parse(args[2:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return fmt.Errorf("存在未知参数")
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		var questions []agent.Question
		var sheet contest.Submission
		if args[0] == "check" {
			if len(data) > contest.MaxBytes {
				return fmt.Errorf("答案超过8 MB")
			}
			if err = agent.DecodeJSON(data, &sheet); err != nil {
				return err
			}
			source := examples
			if *questionPath != "" {
				source, err = os.ReadFile(*questionPath)
				if err != nil {
					return err
				}
			}
			if err = agent.DecodeJSON(source, &questions); err != nil {
				return err
			}
			if *questionPath == "" && sheet.PackageID != "pkg-08e82c8d" {
				return fmt.Errorf("参赛包编号与本包不一致")
			}
			if err = contest.Check(questions, sheet, false); err != nil {
				return err
			}
			fmt.Printf("格式检查通过：%s，共%d题（未执行FOFA搜索）\n", sheet.PackageID, len(sheet.Answers))
			return nil
		}
		if err = agent.DecodeJSON(data, &questions); err != nil {
			return err
		}
		source := templateData
		if *templatePath != "" {
			source, err = os.ReadFile(*templatePath)
			if err != nil {
				return err
			}
		}
		if err = agent.DecodeJSON(source, &sheet); err != nil {
			return err
		}
		if *name != "" {
			sheet.Player = *name
		}
		sheet, err = contest.Generate(questions, sheet)
		if err != nil {
			return err
		}
		encoded, err := contest.Encode(sheet)
		if err != nil {
			return err
		}
		if err = contest.Save(*out, encoded); err != nil {
			return err
		}
		fmt.Printf("已生成 %s：%s，共%d题\n", *out, sheet.PackageID, len(sheet.Answers))
		return nil
	case "serve":
		flags := flag.NewFlagSet("serve", flag.ContinueOnError)
		port := flags.Int("port", 8765, "本地服务端口")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *port < 1 || *port > 65535 {
			return fmt.Errorf("请指定1–65535范围内的端口")
		}
		srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", *port), Handler: server.NewHandler(page, examples, templateData), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
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
