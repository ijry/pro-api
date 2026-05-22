package adapter_test

import (
	"context"
	"testing"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

type fakeAdapter struct{ name string }

func (f *fakeAdapter) Name() string                                              { return f.name }
func (f *fakeAdapter) Capabilities() adapter.Capability                          { return adapter.CapChat }
func (f *fakeAdapter) SupportedModels() []string                                 { return []string{f.name + "-*"} }
func (f *fakeAdapter) Chat(context.Context, *ir.ChatRequest, adapter.Credential) (*ir.ChatResponse, error) {
	return nil, nil
}
func (f *fakeAdapter) ChatStream(context.Context, *ir.ChatRequest, adapter.Credential) (adapter.StreamReader, error) {
	return nil, nil
}
func (f *fakeAdapter) Embed(context.Context, *ir.EmbedRequest, adapter.Credential) (*ir.EmbedResponse, error) {
	return nil, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := adapter.NewRegistry()
	a := &fakeAdapter{name: "openai"}
	r.Register(a)
	got, ok := r.Get("openai")
	if !ok || got.Name() != "openai" {
		t.Fatalf("missing")
	}
	_, ok = r.Get("missing")
	if ok {
		t.Fatalf("expected miss")
	}
}

func TestRegistry_DuplicatePanics(t *testing.T) {
	r := adapter.NewRegistry()
	a := &fakeAdapter{name: "openai"}
	r.Register(a)
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	r.Register(a)
}

func TestRegistry_MustGetUnknownPanics(t *testing.T) {
	r := adapter.NewRegistry()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	r.MustGet("missing")
}

func TestRegistry_ListSorted(t *testing.T) {
	r := adapter.NewRegistry()
	r.Register(&fakeAdapter{name: "openai"})
	r.Register(&fakeAdapter{name: "anthropic"})
	r.Register(&fakeAdapter{name: "gemini"})
	names := r.Names()
	if len(names) != 3 || names[0] != "anthropic" || names[1] != "gemini" || names[2] != "openai" {
		t.Fatalf("expected sorted, got %v", names)
	}
}

func TestCapability_Has(t *testing.T) {
	c := adapter.CapChat | adapter.CapStream | adapter.CapVision
	if !c.Has(adapter.CapChat) || !c.Has(adapter.CapStream) || !c.Has(adapter.CapVision) {
		t.Fatalf("missing caps")
	}
	if c.Has(adapter.CapReasoning) {
		t.Fatalf("unexpected reasoning")
	}
}
