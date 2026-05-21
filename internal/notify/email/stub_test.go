package email

import (
	"context"
	"testing"

	"github.com/ijry/pro-api/internal/audit"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type captureAudit struct{ entries []audit.Entry }

func (c *captureAudit) Log(_ context.Context, e audit.Entry) error {
	c.entries = append(c.entries, e)
	return nil
}

func TestStub_LogsAndAudits(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	ca := &captureAudit{}
	m := NewStub(zap.New(core), ca)
	err := m.Send(context.Background(), Message{
		To: "a@b.com", Subject: "Hi", Body: "code=123456", Tag: "verify_register",
	})
	if err != nil {
		t.Fatal(err)
	}
	if recorded.Len() == 0 {
		t.Fatal("want log entry")
	}
	if len(ca.entries) != 1 || ca.entries[0].Action != "email.send_stub" {
		t.Fatalf("audit wrong: %+v", ca.entries)
	}
}

func TestStub_NilDeps_NoPanic(t *testing.T) {
	m := NewStub(nil, nil)
	if err := m.Send(context.Background(), Message{}); err != nil {
		t.Fatal(err)
	}
}
