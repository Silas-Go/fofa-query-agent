# FOFA 查询助手 · 可运行 MVP

核心流程：自然语言 → 结构化条件 JSON → 条件校验 → 确定性查询生成。使用 Python 标准库，无需安装依赖或配置 API Key；本地网页只是已有的调用入口。

## 启动

需要 Python 3.10 或更新版本。在项目目录运行：

```sh
python3 -m fofa_assistant serve
```

打开 [本地页面](http://127.0.0.1:8765)。按 Ctrl+C 停止；端口占用时加 `--port 8766`。

页面支持单题输入、导入 JSON、加载100道原题、复制查询和导出结果。所有处理留在本机，不调用模型或 FOFA API，也不扫描资产。

## 命令行

```sh
python3 -m fofa_assistant query '搜索 SSH 协议且端口为 22022 的资产。'
python3 -m fofa_assistant answer fixtures/questions.json -o outputs/answers.json
```

输入为包含“题号”“自然语言输入”的 JSON 数组。一次支持1–1000题，题号必须唯一；网页请求大小上限2 MB。结果保留题号、原文、顺序，并增加状态、查询语句、说明和依据。

## 核心流程

- `parser.py` 只生成条件 JSON，未理解的部分保留为 unresolved 节点。
- `validator.py` 校验结构、字段白名单、操作符、参数类型/范围、条件冲突及无法表达的意图。
- `generator.py` 只接收条件 JSON，每次生成前强制校验；失败不输出查询。
- 每条答题结果增加“结构化条件”和“校验结果”，可查看拒绝原因与对应条件路径。

```sh
python3 -m fofa_assistant parse '搜索 SSH 协议且端口为 22022 的资产。' > conditions.json
python3 -m fofa_assistant compile conditions.json
```

结构示例：

```json
{"version":"1.0","condition":{"type":"and","conditions":[{"type":"predicate","field":"protocol","operator":"eq","value":"ssh"},{"type":"predicate","field":"port","operator":"eq","value":22022}]}}
```

端口1000000会返回输入非法；国家同时等于和不等于US会返回条件矛盾；不在白名单中的字段被拒绝。AND/OR 可以通过条件 JSON 明确组合，自然语言混合逻辑建议使用括号。

## 本版范围

保留基础 IP/CIDR、端口、主域/访问主机、常用国家、ASN、TCP/UDP、部分服务协议、标题/正文/响应头/Banner、状态码、JS文件引用以及明确的“且”组合和带括号的“或”组合。规则按输入内容匹配，不按题号返回固定答案。

遇到无法完整理解的句子，整题返回“暂未支持”，不输出缺少条件的半成品查询。自然语言表达的覆盖有限，同义改写也可能暂未支持。

暂缓：联网资料提取、图标哈希、跨平台转换、查询修复、证书、云产品分类、复杂自然语言推理、空值和复杂通配。完整题意保留在原题文件中。

**“暂未支持”表示本版未实现，不代表 FOFA 平台不支持。** 它独立于输入非法、条件矛盾、需要澄清、能力不支持和依赖失败。当前输出约束见 [MVP Schema](docs/mvp-output.schema.json)。

## 本次功能检查

按用户要求只执行了一次核心功能检查：正常条件组合、非法端口拒绝、国家条件冲突、不支持字段拒绝，均通过。查询未在 FOFA 执行，未做完整题集正确率验收。之前的21题生成统计属于旧版，不作为当前版本成绩。

## 项目文件

- `fofa_assistant/engine.py`：串联解析、校验和生成；各层独立模块。
- `fofa_assistant/web.py`、`static/index.html`：本地网页和接口。
- `fixtures/questions.json`：100道原题。
- `outputs/answers.json`：本地运行产物，不上传；运行批量命令可生成当前版本结果。
- [T1 完整目标及验收集](docs/t1/contract.md)、[T2 字段资料](docs/t2/README.md)：后续补齐能力的参考。本 README 的缩减范围为当前交付范围，T1 的95题目标成功不代表本版覆盖。
