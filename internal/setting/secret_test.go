package setting

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeDecryptor struct {
	out string
	err error
}

func (f fakeDecryptor) Decrypt(s string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.out, nil
}

func TestGetSecret_ReturnsPlaintext(t *testing.T) {
	s, _ := newStoreWithDB(t)
	// 存一个 ENC(...) 形态字符串 value(JSON-string-encoded)
	if err := s.Put(context.Background(), "auth.smtp.password", "ENC(v1,nonce,ct)", 1); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetSecret(context.Background(), "auth.smtp.password", fakeDecryptor{out: "hunter2"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "hunter2" {
		t.Fatalf("want hunter2, got %q", got)
	}
}

func TestGetSecret_NotFound(t *testing.T) {
	s, _ := newStoreWithDB(t)
	_, err := s.GetSecret(context.Background(), "missing", fakeDecryptor{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestGetSecret_NotEncrypted_PlainString(t *testing.T) {
	s, _ := newStoreWithDB(t)
	_ = s.Put(context.Background(), "auth.smtp.password", "plaintext", 1)
	_, err := s.GetSecret(context.Background(), "auth.smtp.password", fakeDecryptor{})
	if !errors.Is(err, ErrNotEncrypted) {
		t.Fatalf("want ErrNotEncrypted, got %v", err)
	}
}

func TestGetSecret_NotEncrypted_NonString(t *testing.T) {
	s, _ := newStoreWithDB(t)
	_ = s.Put(context.Background(), "flag", true, 1)
	_, err := s.GetSecret(context.Background(), "flag", fakeDecryptor{})
	if !errors.Is(err, ErrNotEncrypted) {
		t.Fatalf("want ErrNotEncrypted, got %v", err)
	}
}

func TestGetSecret_DecryptError(t *testing.T) {
	s, _ := newStoreWithDB(t)
	_ = s.Put(context.Background(), "auth.smtp.password", "ENC(v1,n,c)", 1)
	_, err := s.GetSecret(context.Background(), "auth.smtp.password", fakeDecryptor{err: errors.New("bad nonce")})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestListAll_ReturnsAllRowsOrderedByKey(t *testing.T) {
	s, _ := newStoreWithDB(t)
	now := time.Now().UTC()
	// 直接 db.Create 避免 Put 的 redis del 干扰
	_ = s.db.Create(&Setting{Key: "z", Value: []byte(`1`), UpdatedAt: now}).Error
	_ = s.db.Create(&Setting{Key: "a", Value: []byte(`2`), UpdatedAt: now}).Error
	_ = s.db.Create(&Setting{Key: "m", Value: []byte(`3`), UpdatedAt: now}).Error

	rows, err := s.ListAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("want 3, got %d", len(rows))
	}
	got := []string{rows[0].Key, rows[1].Key, rows[2].Key}
	want := []string{"a", "m", "z"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order: want %v, got %v", want, got)
		}
	}
}

func TestListAll_EmptyTable(t *testing.T) {
	s, _ := newStoreWithDB(t)
	rows, err := s.ListAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("want empty, got %d", len(rows))
	}
}
