import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

from .engine import answer_batch
from . import __version__

ROOT = Path(__file__).resolve().parents[1]


class Handler(BaseHTTPRequestHandler):
    def reply(self, status, body, content_type="application/json; charset=utf-8"):
        if not isinstance(body, bytes):
            body = json.dumps(body, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store")
        self.send_header("X-Content-Type-Options", "nosniff")
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        if self.path == "/":
            self.reply(200, (Path(__file__).parent / "static/index.html").read_bytes(), "text/html; charset=utf-8")
        elif self.path == "/api/examples":
            self.reply(200, json.loads((ROOT / "fixtures/questions.json").read_text()))
        elif self.path == "/health":
            self.reply(200, {"status": "ok", "version": __version__})
        else:
            self.reply(404, {"error": "页面不存在。"})

    def do_POST(self):
        if self.path != "/api/answer":
            return self.reply(404, {"error": "接口不存在。"})
        if self.headers.get("Origin") and self.headers["Origin"] != f"http://{self.headers.get('Host')}":
            return self.reply(403, {"error": "只允许本地页面请求。"})
        try:
            size = int(self.headers.get("Content-Length", "0"))
            if not 0 < size <= 2_000_000:
                return self.reply(413, {"error": "输入为空或超过 2 MB。"})
            self.connection.settimeout(10)
            data = json.loads(self.rfile.read(size).decode("utf-8"))
            self.reply(200, {"results": answer_batch(data)})
        except (ValueError, OSError) as exc:
            self.reply(400, {"error": str(exc)})

    def log_message(self, fmt, *args):
        pass


def serve(port=8765):
    server = ThreadingHTTPServer(("127.0.0.1", port), Handler)
    print(f"FOFA 查询助手已启动：http://127.0.0.1:{server.server_port}", flush=True)
    print("按 Ctrl+C 停止。", flush=True)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()
