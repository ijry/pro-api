package adapterreg_test

import (
	"testing"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/adapterreg"
)

func TestWireAdaptersRegistersGrokProviders(t *testing.T) {
	reg := adapter.NewRegistry()
	adapterreg.WireAdapters(reg, nil)
	for _, name := range []string{"grok-build", "grok-web"} {
		a, ok := reg.Get(name)
		if !ok {
			t.Fatalf("missing adapter %q; names=%v", name, reg.Names())
		}
		if a.Name() != name {
			t.Fatalf("adapter name = %q, want %q", a.Name(), name)
		}
	}
}
