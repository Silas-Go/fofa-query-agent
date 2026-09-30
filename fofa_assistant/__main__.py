import argparse
import json
import sys
from collections import Counter
from pathlib import Path

from .engine import answer_batch


def main():
    parser = argparse.ArgumentParser(description="FOFA 查询助手：本地规则 MVP，无外部依赖")
    commands = parser.add_subparsers(dest="command", required=True)
    batch = commands.add_parser("answer", help="处理 JSON 题目文件")
    batch.add_argument("input", type=Path)
    batch.add_argument("-o", "--output", type=Path)
    one = commands.add_parser("query", help="处理一句自然语言")
    one.add_argument("text")
    parse = commands.add_parser("parse", help="只生成结构化条件 JSON")
    parse.add_argument("text")
    compile_cmd = commands.add_parser("compile", help="校验条件 JSON 并生成查询")
    compile_cmd.add_argument("input", type=Path)
    serve = commands.add_parser("serve", help="启动本地网页")
    serve.add_argument("--port", type=int, default=8765)
    args = parser.parse_args()
    if args.command == "serve":
        from .web import serve as start
        try:
            start(args.port)
        except (OSError, ValueError) as exc:
            parser.exit(2, f"无法启动：{exc}\n")
        return
    try:
        if args.command == "parse":
            from .parser import parse_conditions
            print(json.dumps(parse_conditions(args.text), ensure_ascii=False, indent=2))
            return
        if args.command == "compile":
            from .generator import generate_query
            document = json.loads(args.input.read_text(encoding="utf-8-sig"))
            print(generate_query(document))
            return
        data = [{"题号": "Q001", "自然语言输入": args.text}] if args.command == "query" else json.loads(args.input.read_text(encoding="utf-8-sig"))
        results = answer_batch(data)
        text = json.dumps(results, ensure_ascii=False, indent=2) + "\n"
        target = getattr(args, "output", None)
        if target:
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text(text, encoding="utf-8")
            print(f"已保存 {len(results)} 条结果至 {target}", file=sys.stderr)
            print(" / ".join(f"{k} {v}" for k, v in Counter(r['状态'] for r in results).items()), file=sys.stderr)
        else:
            print(text, end="")
    except (OSError, ValueError) as exc:
        parser.exit(2, f"处理失败：{exc}\n")


if __name__ == "__main__":
    main()
