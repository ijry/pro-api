package adapter_test

import (
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/util/tokenize"
)

func TestNewHTTPClient_TimeoutDefaults(t *testing.T) {
	c := adapter.NewHTTPClient(adapter.ClientConfig{Provider: "openai", Timeout: 30 * time.Second})
	if c.Timeout != 30*time.Second {
		t.Fatalf("timeout: %v", c.Timeout)
	}
	if c.CheckRedirect == nil {
		t.Fatalf("CheckRedirect should be set")
	}
}

func TestNewHTTPClient_ZeroTimeoutForStream(t *testing.T) {
	c := adapter.NewHTTPClient(adapter.ClientConfig{Provider: "openai", Timeout: 0})
	if c.Timeout != 0 {
		t.Fatalf("expected zero timeout (stream mode)")
	}
}

func TestRegisterTokenizers_RegistersOpenAIAndOthers(t *testing.T) {
	reg := tokenize.NewRegistry(tokenize.NewApproximate(4))
	adapter.RegisterTokenizers(reg)
	// gpt-4o → tiktoken or approximate
	if reg.For("gpt-4o-mini") == nil {
		t.Fatalf("gpt-4o tokenizer not registered")
	}
	if reg.For("claude-3-5-sonnet-20241022") == nil {
		t.Fatalf("claude tokenizer not registered")
	}
	if reg.For("gemini-1.5-pro") == nil {
		t.Fatalf("gemini tokenizer not registered")
	}
	if reg.For("deepseek-chat") == nil {
		t.Fatalf("deepseek tokenizer not registered")
	}
}
