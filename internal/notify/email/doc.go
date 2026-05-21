// Package email 提供发件 Mailer 接口与 M1 stub 实现。
//
// M1 仅有 Stub:把邮件内容打到 log + audit,SMTP 实现 M2 再补一个 smtp.go。
package email
