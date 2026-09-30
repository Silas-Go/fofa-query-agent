# T2：FOFA 能力字典首版

已整理 **75 个官方搜索语法条目**，可供后续解析和生成使用。来源是 [FOFA 官方页面](https://fofa.info/)实际加载的[语法资源](https://v5static.fofa.info/_nuxt/Bw0mZRLA.js)，核对日期 2026-09-30。包括字段、函数和组合条目，并非75种不同资产属性。

- [字段速查](fields.md)：阅读版。
- [能力字典](capabilities.json)：字段、操作符、示例、类型约定、来源摘要和未确认项。
- [逐题字段映射](question-fields.json)：100题对应的字段及需要特殊处理的题目。

## 直接用于实现的结论

| 用户概念 | 查询字段/表达 |
|---|---|
| 主域 / 访问主机 | `domain` / `host` |
| 服务协议 / TCP、UDP | `protocol` / `base_protocol` |
| 服务响应 / HTTP头 / 网页正文 | `banner` / `header` / `body` |
| 引用 JS 文件 | `js_name`，M031 优先用此字段 |
| 产品 / 产品门类 / 产品版本 | `product` / `category` / `product.version` |
| 已绑定域名 | `is_domain=true` |
| 网站类记录 | `type="subdomain"` |
| 是否云资产 / 云服务商 | `is_cloud` / `cloud_name` |
| 未过期 | `cert.is_expired=false` |
| 有效 | `cert.is_valid=true` |
| 颁发者与持有者不匹配 | `cert.is_equal=false`，对应 M047 |
| 证书与域名不匹配 | `cert.is_match=false`，对应 M048 |
| 证书到期时间范围 | `cert.not_after.after` / `cert.not_after.before` |
| 资产更新时间范围 | `after` / `before` |

这些来自搜索语法表。API返回字段中的 `product_category`、`lastupdatetime`、`cert.not_after` 不能直接替代上面的搜索字段。

`=` 与 `==` 分开处理；逻辑组合使用显式括号。模糊操作符是 `*=`，不能把它和普通包含匹配混为一谈。`app` 在此次官方表中没有标注支持 `!=`，排除应用时不能机械生成该表达。

M100 序列号应保持大整数精度。官方 `cert.sn` 示例采用十进制数字字符串；题目给定数值转换为 `132577581667764526475870473477991557894`。颁发者通用名称对应 `cert.issuer.cn`，不是组织字段。

## 留到生成阶段集中处理的事项

1. **具体枚举**：亚马逊三类服务、华为云、DigitalOcean、海康摄像头、HFS、Grafana、WordPress、Apache HTTP Server、邮件系统、中间件、SOCKS v5，以及地域别名。字段已知不代表这些值已确认，不能套用展示名称猜标签。
2. **匹配细节**：各字段的 `==`、空值、非空、通配符单字符、域名标签边界；尤其 M023、M037、M042、M076、M081、M098。
3. **特定限制**：M024 独立版本搜索；M082 秒级日期边界；M038/M041/M042/M058 等题中的“证书含域名”是否等同 `cert.domain` 的持有者根域语义。
4. **跨平台与外部材料**：M083/M086/M087 等价转换；M097实际图标及哈希。不能用未确认的近似表达冒充精确答案。

以上记录为待确认，没有提前判为“平台不支持”。此轮按用户要求不做细致测试或真实查询；最终功能检查时集中核对。先用已确认字典继续 T3/T4，避免在资料整理阶段停住。
