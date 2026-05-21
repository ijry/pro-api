package email

import "context"

// Message 是一封邮件的归一化结构。
type Message struct {
	To       string
	Subject  string
	Body     string // 纯文本
	BodyHTML string // 可选,Stub 忽略
	Tag      string // verify_register / verify_login / password_reset 等
}

// Mailer 是发件抽象。
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}
