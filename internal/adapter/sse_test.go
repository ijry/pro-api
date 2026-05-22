package adapter_test

import (
	"io"
	"strings"
	"testing"

	"github.com/ijry/pro-api/internal/adapter"
)

func TestSSEReader_SingleDataEvent(t *testing.T) {
	r := adapter.NewSSEReader(strings.NewReader("data: hello\n\n"))
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if ev.Data != "hello" {
		t.Fatalf("data: %q", ev.Data)
	}
}

func TestSSEReader_MultiLineData(t *testing.T) {
	r := adapter.NewSSEReader(strings.NewReader("data: a\ndata: b\n\n"))
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if ev.Data != "a\nb" {
		t.Fatalf("data: %q", ev.Data)
	}
}

func TestSSEReader_CommentSkipped(t *testing.T) {
	r := adapter.NewSSEReader(strings.NewReader(": comment\ndata: hi\n\n"))
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if ev.Data != "hi" {
		t.Fatalf("data: %q", ev.Data)
	}
}

func TestSSEReader_NamedEventAndID(t *testing.T) {
	r := adapter.NewSSEReader(strings.NewReader("event: ping\nid: 42\ndata: hi\n\n"))
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if ev.Event != "ping" || ev.ID != "42" || ev.Data != "hi" {
		t.Fatalf("bad event: %+v", ev)
	}
}

func TestSSEReader_MultipleEvents(t *testing.T) {
	r := adapter.NewSSEReader(strings.NewReader("data: 1\n\ndata: 2\n\n"))
	if ev, err := r.Next(); err != nil || ev.Data != "1" {
		t.Fatalf("first: %+v, err=%v", ev, err)
	}
	if ev, err := r.Next(); err != nil || ev.Data != "2" {
		t.Fatalf("second: %+v, err=%v", ev, err)
	}
}

func TestSSEReader_EOFAfterFinalEvent(t *testing.T) {
	// no trailing blank line — emit final event on EOF then EOF
	r := adapter.NewSSEReader(strings.NewReader("data: tail"))
	if ev, err := r.Next(); err != nil || ev.Data != "tail" {
		t.Fatalf("first: %+v, err=%v", ev, err)
	}
	if _, err := r.Next(); err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
}
