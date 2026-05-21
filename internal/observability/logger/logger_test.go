package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNew_InvalidLevel(t *testing.T) {
	if _, err := New("trace", "json", nil); err == nil {
		t.Fatal("want error for unknown level")
	}
}

func TestNew_InvalidFormat(t *testing.T) {
	if _, err := New("info", "yaml", nil); err == nil {
		t.Fatal("want error for unknown format")
	}
}

func TestNew_JSON_Output(t *testing.T) {
	var buf bytes.Buffer
	log, err := New("info", "json", zapcore.AddSync(&buf))
	if err != nil {
		t.Fatal(err)
	}
	log.Info("hello", zap.String("k", "v"))
	_ = log.Sync()

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected json: %v\n%s", err, buf.String())
	}
	if entry["msg"] != "hello" || entry["k"] != "v" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
}

func TestNew_LevelFilter(t *testing.T) {
	var buf bytes.Buffer
	log, _ := New("warn", "json", zapcore.AddSync(&buf))
	log.Info("ignored")
	log.Warn("kept")
	_ = log.Sync()
	if bytes.Contains(buf.Bytes(), []byte("ignored")) {
		t.Fatal("info should be filtered out at warn level")
	}
	if !bytes.Contains(buf.Bytes(), []byte("kept")) {
		t.Fatal("warn message missing")
	}
}
