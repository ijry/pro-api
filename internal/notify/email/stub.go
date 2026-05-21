package email

import (
	"context"
	"encoding/json"

	"github.com/ijry/pro-api/internal/audit"
	"go.uber.org/zap"
)

// stub 把邮件打到 log + audit,便于开发调试。
type stub struct {
	log   *zap.Logger
	audit audit.Logger
}

// NewStub 构造 stub Mailer。
func NewStub(log *zap.Logger, a audit.Logger) Mailer {
	if log == nil {
		log = zap.NewNop()
	}
	if a == nil {
		a = audit.NewNoop()
	}
	return &stub{log: log, audit: a}
}

// Send 实现 Mailer。
func (s *stub) Send(ctx context.Context, m Message) error {
	s.log.Info("[email stub] send",
		zap.String("to", m.To),
		zap.String("subject", m.Subject),
		zap.String("tag", m.Tag),
		zap.String("body", m.Body),
	)
	after, _ := json.Marshal(map[string]any{
		"to":      m.To,
		"subject": m.Subject,
		"tag":     m.Tag,
	})
	_ = s.audit.Log(ctx, audit.Entry{
		Action:     "email.send_stub",
		TargetType: "email",
		After:      after,
	})
	return nil
}
