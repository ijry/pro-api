# 项目介绍

> 💡 **想直接看看?** <a href="/admin-demo/index.html" target="_blank">管理后台演示</a> · <a href="/user-demo/index.html" target="_blank">用户中心演示</a>(纯前端 mock,无需部署后端)

**proapi** 是一个自主实现的大模型 API 中转 / 网关系统,定位为开源项目 + 对外 SaaS + 企业平台三位一体。

## 核心能力

- **三协议入口互转**:OpenAI、Anthropic、Gemini 任一入口可路由到任一后端
- **18+ 上游适配**:覆盖海内外主流大模型厂商
- **企业级账号体系**:OAuth(GitHub/Google/微信/飞书/钉钉/Discord)+ SSO(OIDC/SAML/LDAP/CAS)
- **完整计费**:Lua 原子扣费、模型倍率、分组倍率、部门预算
- **可观测**:Prometheus 内置 + 可选 OpenTelemetry
- **高可定制**:适配器、支付、OAuth 全部接口化,二次开发友好

## 适合谁

- **个人 / 小团队**:聚合多家模型,避免多套 API key 管理
- **SaaS 运营者**:开箱即用的计费、邀请、充值
- **企业 IT**:可对接现有 SSO,部门级预算,审计合规

## 不做什么

- 不自营模型推理
- 不做 Agent / 工作流编排
- 不做 RAG / 向量数据库
