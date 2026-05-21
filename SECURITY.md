# 安全策略

## 漏洞披露

发现安全问题请发邮件至 `security@proapi.example`(M2 上线前替换为真实邮箱),不要直接开 issue。

90 天内会披露,补丁先于公告。

## 加密密钥

- 应用层加密用 AES-256-GCM
- Master key 从 `PROAPI_MASTER_KEY` 环境变量读取,32 字节(base64)
- 丢失 master key → 已加密数据(渠道凭证、OAuth secret 等)不可恢复,运维需备份

## 安全审查

每个 release 走一次 `make audit` 与依赖扫描:

- `govulncheck ./...`
- `pnpm audit`
- `trivy image` 扫容器
