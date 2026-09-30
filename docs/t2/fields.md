# FOFA 字段速查

从[官方页面](https://fofa.info/)公开语法表提取，2026-09-30。完整数据及来源摘要见 [capabilities.json](capabilities.json)。

表中的操作符为逐字段公开标记。`==`另见全局表，不能据此推断所有字段支持。所有条目均未执行查询；空值和缺失值行为尚未证明。

|语义|字段/条目|已列操作符|官方示例|
|---|---|---|---|
|资产IP或CIDR|`ip`|`=`、`!=`|`ip="1.1.1.1"`|
|服务端口|`port`|`=`、`!=`、`*=`|`port="6379"`|
|资产根域|`domain`|`=`、`!=`、`*=`|`domain="qq.com"`|
|访问主机名|`host`|`=`、`!=`、`*=`|`host=".fofa.info"`|
|操作系统识别|`os`|`=`、`!=`、`*=`|`os="centos"`|
|服务器标识|`server`|`=`、`!=`、`*=`|`server="Microsoft-IIS/10"`|
|自治系统号|`asn`|`=`、`!=`、`*=`|`asn="19551"`|
|网络组织|`org`|`=`、`!=`、`*=`|`org="LLC Baxet"`|
|是否绑定域名|`is_domain`|`=`|`is_domain=true`|
|是否IPv6|`is_ipv6`|`=`|`is_ipv6=true`|
|FOFA应用识别规则|`app`|`=`|`app="Microsoft-Exchange"`|
|站点指纹|`fid`|`=`、`!=`|`fid="sSXXGNUO2FefBTcCLIT/2Q=="`|
|FOFA产品标签|`product`|`=`、`!=`|`product="NGINX"`|
|产品版本|`product.version`|`=`、`!=`|`product="Roundcube-Webmail" && product.version="1.6.10"`|
|产品门类|`category`|`=`、`!=`|`category="服务"`|
|资产记录类型|`type`|`=`|`type="service"`|
|云服务商|`cloud_name`|`=`、`!=`、`*=`|`cloud_name="Aliyundun"`|
|是否云资产|`is_cloud`|`=`|`is_cloud=true`|
|欺诈标记|`is_fraud`|`=`|`is_fraud=true`|
|蜜罐标记|`is_honeypot`|`=`|`is_honeypot=true`|
|服务协议|`protocol`|`=`、`!=`、`*=`|`protocol="quic"`|
|服务响应原文|`banner`|`=`、`!=`|`banner="users"`|
|服务响应哈希|`banner_hash`|`=`、`!=`|`banner_hash="7330105010150477363"`|
|服务响应结构指纹|`banner_fid`|`=`、`!=`|`banner_fid="zRpqmn0FXQRjZpH8MjMX55zpMy9SgsW8"`|
|传输协议|`base_protocol`|`=`、`!=`|`base_protocol="udp"`|
|网页标题|`title`|`=`、`!=`、`*=`|`title="beijing"`|
|HTTP响应头|`header`|`=`、`!=`|`header="elastic"`|
|响应头哈希|`header_hash`|`=`、`!=`、`*=`|`header_hash="1258854265"`|
|网页正文|`body`|`=`、`!=`|`body="网络空间测绘"`|
|正文哈希|`body_hash`|`=`、`!=`|`body_hash="-2090962452"`|
|引用的JS文件名|`js_name`|`=`、`!=`、`*=`|`js_name="js/jquery.js"`|
|引用JS文件的MD5|`js_md5`|`=`、`!=`、`*=`|`js_md5="82ac3f14327a8b7ba49baa208d4eaa15"`|
|CNAME记录|`cname`|`=`、`!=`、`*=`|`cname="customers.spektrix.com"`|
|CNAME根域|`cname_domain`|`=`、`!=`、`*=`|`cname_domain="siteforce.com"`|
|图标哈希|`icon_hash`|`=`、`!=`|`icon_hash="-247388890"`|
|网站响应码|`status_code`|`=`、`!=`|`status_code="402"`|
|备案信息|`icp`|`=`、`!=`、`*=`|`icp="京ICP证030173号"`|
|前端SDK哈希|`sdk_hash`|`=`、`!=`|`sdk_hash="Are3qNnP2Eqn7q5kAoUO3l+w3mgVIytO"`|
|国家或地区代码|`country`|`=`、`!=`|`country="CN"`|
|省州级地域|`region`|`=`、`!=`|`region="Zhejiang"`|
|城市|`city`|`=`、`!=`|`city="Hangzhou"`|
|证书全文|`cert`|`=`、`!=`|`cert="baidu"`|
|证书持有者信息|`cert.subject`|`=`、`!=`、`*=`|`cert.subject="Oracle Corporation"`|
|证书颁发者信息|`cert.issuer`|`=`、`!=`、`*=`|`cert.issuer="DigiCert"`|
|证书持有者组织|`cert.subject.org`|`=`、`!=`、`*=`|`cert.subject.org="Oracle Corporation"`|
|证书持有者通用名称|`cert.subject.cn`|`=`、`!=`、`*=`|`cert.subject.cn="baidu.com"`|
|证书颁发者组织|`cert.issuer.org`|`=`、`!=`、`*=`|`cert.issuer.org="cPanel, Inc."`|
|证书颁发者通用名称|`cert.issuer.cn`|`=`、`!=`、`*=`|`cert.issuer.cn="Synology Inc. CA"`|
|证书持有者根域|`cert.domain`|`=`、`!=`、`*=`|`cert.domain="huawei.com"`|
|证书颁发者与持有者是否匹配|`cert.is_equal`|`=`|`cert.is_equal=true`|
|证书是否有效|`cert.is_valid`|`=`|`cert.is_valid=true`|
|证书与域名是否匹配|`cert.is_match`|`=`|`cert.is_match=true`|
|证书是否过期|`cert.is_expired`|`=`|`cert.is_expired=true`|
|JARM指纹|`jarm`|`=`、`!=`、`*=`|`jarm="2ad2ad0002ad2ad22c2ad2ad2ad2ad2eac92ec34bcc0cf7520e97547f83e81"`|
|TLS版本|`tls.version`|`=`、`!=`|`tls.version="TLS 1.3"`|
|JA3S指纹|`tls.ja3s`|`=`、`!=`、`*=`|`tls.ja3s="15af977ce25de452b96affa2addb1036"`|
|证书序列号|`cert.sn`|`=`、`!=`|`cert.sn="356078156165546797850343536942784588840297"`|
|证书到期日晚于|`cert.not_after.after`|`=`|`cert.not_after.after="2025-03-01"`|
|证书到期日早于|`cert.not_after.before`|`=`|`cert.not_after.before="2025-03-01"`|
|证书生效日晚于|`cert.not_before.after`|`=`|`cert.not_before.after="2025-03-01"`|
|证书生效日早于|`cert.not_before.before`|`=`|`cert.not_before.before="2025-03-01"`|
|资产更新时间晚于|`after`|`=`|`after="2023-01-01"`|
|资产更新时间早于|`before`|`=`|`before="2023-12-01"`|
|资产更新时间区间组合|`after&before`|`=`|`after="2023-01-01" && before="2023-12-01"`|
|独立IP聚合过滤|`ip_filter()`|`=`|`ip_filter(banner="SSH-2.0-OpenSSH_6.7p2") && ip_filter(icon_hash="-1057022626")`|
|独立IP聚合排除|`ip_exclude()`|`=`|`ip_filter(banner="SSH-2.0-OpenSSH_6.7p2" && asn="3462") && ip_exclude(title="EdgeOS")`|
|独立IP端口数量|`port_size`|`=`、`!=`|`port_size="6"`|
|独立IP端口数量大于|`port_size_gt`|`=`|`port_size_gt="6"`|
|独立IP端口数量小于|`port_size_lt`|`=`|`port_size_lt="12"`|
|独立IP开放端口集合|`ip_ports`|`=`|`ip_ports="80,161"`|
|独立IP国家|`ip_country`|`=`|`ip_country="CN"`|
|独立IP省州|`ip_region`|`=`|`ip_region="Zhejiang"`|
|独立IP城市|`ip_city`|`=`|`ip_city="Hangzhou"`|
|独立IP更新晚于|`ip_after`|`=`|`ip_after="2021-03-18"`|
|独立IP更新早于|`ip_before`|`=`|`ip_before="2019-09-09"`|
