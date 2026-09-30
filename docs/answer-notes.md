# 本包作答记录

这份记录不进入正式答卷。当前条件树由 Agent 逐题解释生成，经过本地结构与参数校验；没有官方参考答案，也未执行 FOFA 搜索。95条查询、5条拒绝仅表示输出构成。

## 来源和计算

- FOFA字段依据：[官方页面](https://fofa.info/)及其[公开语法资源](https://v5static.fofa.info/_nuxt/Bw0mZRLA.js)，已整理在 `docs/t2/`。
- M096：[Jenkins TCP Agent Listener Port](https://www.jenkins.io/doc/book/security/services/) 的示例包含四个符合前缀条件的字段：Jenkins-Agent-Protocols、Jenkins-Version、Jenkins-Session、Remoting-Minimum-Version。条件树对 body/banner 共8个分支做OR；未限制端口或版本。
- M097：读取 [Octodex首页](https://octodex.github.com/) 的 icon 链接，实际下载 `/favicon.ico`；6518字节，SHA256为 `20c67acbdf77f66d5f959a91818950c8873455151a123ea36641ccfda7a52cc3`。按照 [FOFA官方Go客户端](https://github.com/FofaInfo/GoFOFA/blob/main/iconhash.go) 的Base64每76字符换行、末尾换行、MurmurHash3 32位有符号数算法，计算为 `-1493334710`。排除原站点 host。此值对应2026-09-30获取的图标，比赛环境的历史图标可能不同。
- M099：[Apache官方漏洞说明](https://httpd.apache.org/security/vulnerabilities_24.html) 将 CVE-2021-41773 归于 Apache HTTP Server。题目只问对应产品，因此查询不添加漏洞利用、版本或可利用性假设。
- M100：十六进制证书序列号转为十进制字符串 `132577581667764526475870473477991557894`，避免浮点精度丢失；颁发者通用名称使用 `cert.issuer.cn`。

## 需要执行环境核验的选择

以下不是已验证的标准答案，保留为复核清单：

- M007/M008 使用 Gangwon-do、Bavaria、Munich 地域名称；数据库实际规范名可能不同。
- M010/M011/M022/M025/M054/M066/M070/M072/M099 的产品、分类和云厂商标签按题意选择，未通过 FOFA 枚举接口核对。尤其 WAF 产品标签是否完整覆盖题意需检查。
- M023/M052/M072/M076/M098 使用 `*=` 通配；单字符、完整主机边界和账户权限需核验。
- M034/M100 的完整证书名称用 `==`；字段级精确匹配兼容性需核验。
- M037/M076/M081 使用空/非空语义；缺失字段与空值的实际执行差异需核验。
- M038/M041/M042/M058/M074/M082 对“证书包含域名”保留全文匹配，与持有者根域字段分开；全文匹配可能命中其他证书文字。
- M083 的 Shodan自由文本映射为banner；M086 的域名包含映射为host；M087 的头部键值映射为配对文本。跨平台索引和规范化可能不同。
- M082 保留秒级时间，不擅自截断为日期；执行环境是否支持秒级边界未验证。
- M098 根据输入域名形态概括为 `*.gov.cn.*`，这是结构特征推断，不证明匹配网站有恶意行为。

M091–M095分别为非法端口、非法IP、同记录国家矛盾、缺乏可检索定义、缺少运行时监控数据，按公告填写固定拒绝句。未实现或外部资料读取失败不能归入这一组；生成器遇到这类缺口会阻止导出。
