# FOFA 查询 Agent · Go MVP

用 Go 标准库实现：**自然语言 → 结构化条件 JSON → 校验 → FOFA 查询语句**。无需 API Key 或第三方依赖，查询生成器只接收结构化条件。

## 运行

需要 Go 1.22 或更新版本：

```sh
go run . serve
```

打开 [本地页面](http://127.0.0.1:8765)，支持单题、导入题目和导出结果。端口占用时加 `--port 8766`，Ctrl+C 停止。页面和原题嵌入二进制，编译后可独立运行，无需 Python。

```sh
go build -o bin/fofa-query-agent .
./bin/fofa-query-agent serve
```

## 命令行

```sh
go run . query '搜索 SSH 协议且端口为 22022 的资产。'
go run . answer fixtures/questions.json -o outputs/answers.json
go run . parse '搜索 SSH 协议且端口为 22022 的资产。' > conditions.json
go run . compile conditions.json
```

输入题目为包含“题号”“自然语言输入”的 JSON 数组，支持1–1000题、题号唯一。网页请求上限2 MB。结果保留原文和顺序，并给出查询、结构化条件、校验结果及拒绝原因。

条件示例：

```json
{"version":"1.0","condition":{"type":"and","conditions":[{"type":"predicate","field":"protocol","operator":"eq","value":"ssh"},{"type":"predicate","field":"port","operator":"eq","value":22022}]}}
```

生成：`(protocol="ssh" && port="22022")`。

## 模块和边界

- `internal/agent/parser.go`：规则解析为条件树，无法理解的部分保留为 `unresolved`。
- `internal/agent/validator.go`：检查结构、字段白名单、操作符、类型、范围、条件矛盾及无法表达的意图。
- `internal/agent/generator.go`：强制校验后生成查询，失败不输出查询。
- `internal/agent/engine.go`：串联流程；`internal/server/` 和 `main.go` 提供网页接口与命令行入口。

当前支持基础 IP/CIDR、端口、主域/主机、部分国家和协议、ASN、标题/正文/响应头/Banner、状态码、JS文件引用及 AND/OR 组合。自然语言混合逻辑建议使用括号；可直接用条件 JSON 指定分组。

端口越界、IP非法、国家条件矛盾、不支持字段都会被拒绝。规则覆盖有限，同义改写可能返回“暂未支持”；任何未理解条件都阻止整题生成。它表示本版未实现，不代表 FOFA 不支持。

所有处理在本地完成，不调用模型、FOFA API或执行扫描。暂缓联网资料提取、图标哈希、跨平台转换、查询修复、证书、云产品分类和复杂自然语言推理。冲突校验覆盖已实现的明确矛盾，不能证明所有查询都有真实匹配资产。

## 功能检查

```sh
go test ./...
```

包含正常转换、AND/OR组合、非法输入、冲突、不支持字段、无法观测的意图及网页接口。未做100题正确率验收，也未执行真实FOFA搜索。

## 题目与资料

- [100道原题](fixtures/questions.json)
- [当前输出约束](docs/mvp-output.schema.json)
- [T1完整目标和验收集](docs/t1/contract.md)、[T2字段资料](docs/t2/README.md)：保留题意和后续扩展参考；其中的目标覆盖率不代表当前MVP成绩。

批量输出写入本地 `outputs/`，不上传仓库。
