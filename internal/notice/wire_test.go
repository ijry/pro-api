package notice

import (
	"testing"

	"github.com/ijry/pro-api/internal/app"
)

func TestWireNotice_RejectsNilApp(t *testing.T) {
	if err := WireNotice(nil); err == nil {
		t.Fatal("want error")
	}
}

func TestWireNotice_RejectsMissingDB(t *testing.T) {
	a := &app.Application{}
	if err := WireNotice(a); err == nil {
		t.Fatal("want error for missing DB")
	}
}

func TestServiceFrom_NilApp(t *testing.T) {
	if ServiceFrom(nil) != nil {
		t.Fatal("want nil")
	}
}

func TestServiceFrom_WrongType(t *testing.T) {
	a := &app.Application{NoticeSvc: "not a service"}
	if ServiceFrom(a) != nil {
		t.Fatal("want nil for wrong type")
	}
}
