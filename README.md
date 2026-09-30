# FOFA 查询答卷生成器（Go）

按参赛包要求生成可上传的 `答案.json`。当前包：`pkg-08e82c8d`，选手：张致远，共100题。使用 Go 1.22+ 标准库，无第三方运行依赖。

**本包答卷：[submissions/答案.json](submissions/答案.json)**。已完成本地格式与完整性检查，未在比赛网站提交或获得评分。

## 生成与检查

```sh
go run . answer fixtures/questions.json -o submissions/答案.json
go run . check submissions/答案.json
```

输出只含比赛允许的内容：

```json
{
  "选手名称": "张致远",
  "参赛包编号": "pkg-08e82c8d",
  "答案": [
    {"题号": "M001-S009", "查询语句": "ip=\"20.247.40.92\""}
  ]
}
```

上面仅演示一条；实际答卷必须覆盖全部100题。非法输入、矛盾或确实不能直接表达的需求填写固定句子：`该需求不能直接转换为FOFA搜索语句`。

缺题、重复题号、未知题号、空答案、缺少选手名称或包编号都会被拦截。输出为 UTF-8 JSON，限制8 MB；输入允许 BOM。导出前完成检查，再原子写入文件，避免留下半份答卷。

使用其他参赛包时必须提供对应模板，保持其中包编号：

```sh
go run . answer 新题目.json --template 新答案模板.json --name 张致远 -o 答案.json
go run . check 答案.json --questions 新题目.json
```

自定义包的检查只核验格式与题号，包编号应与发放文件自行对照。未实现的需求会阻止生成，不会冒充“不能转换”。

## 核心流程

`自然语言 → 结构化条件 JSON → 条件校验 → 确定性查询生成 → 比赛答卷`

- `internal/agent/parser.go`：基础规则解析。
- `internal/agent/reviewed.json`：Agent 对当前100道题逐题审阅后保存的条件树。按完整原文匹配，不按题号命中；这是本包的作答资料，**不是官方参考答案，也不代表能泛化到任意新题**。
- `internal/agent/validator.go`：字段、类型、参数范围、组合结构和已实现的矛盾检查。
- `internal/agent/generator.go`：只接受条件树，校验通过后生成查询；不直接拼接模型输出。
- `internal/contest/`：保留模板元数据、核对全部题号、输出正式答卷。状态、说明、依据等调试信息不会混入答案文件。

当前本包生成95条查询，5题使用固定拒绝句。**生成数量不是正确率。** 部分产品/地域枚举、空值和通配符、跨平台等价性仍需用比赛环境核验；没有 FOFA 执行验证或评分结果。[作答依据与具体不确定项](docs/answer-notes.md)。

## 调试与原有页面

```sh
go run . query '搜索 SSH 协议且端口为 22022 的资产。'
go run . parse '搜索 SSH 协议且端口为 22022 的资产。' > conditions.json
go run . compile conditions.json
go run . serve
```

打开 [本地页面](http://127.0.0.1:8765)。加载100道原题后，“导出比赛答卷”使用与命令行相同的校验和封装。单题结果可以查看和复制，但不能作为本包完整答卷导出。服务仅监听本机，不执行资产扫描。

```sh
go test ./...
go build -o bin/fofa-query-agent .
```

检查覆盖正常转换、拒绝转换、条件组合、条件树往返序列化、整包导出及缺题/重复/未知题拦截。无需启动外部服务。

## 文件

- [100道原题](fixtures/questions.json)、[答案模板](fixtures/answer-template.json)、[参赛包信息](fixtures/package.json)
- [比赛提交约束](docs/contest-rules.md)、[提交格式 Schema](docs/submission.schema.json)
- [FOFA字段资料](docs/t2/README.md)

`docs/t1/` 与旧 `docs/mvp-output.schema.json` 保留开发阶段资料；其中的内部状态和调试格式不属于比赛提交协议。本 README 与比赛提交约束是当前交付说明。
