package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/group"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeIDGen struct{ n int64 }

func (f *fakeIDGen) Generate() int64 { f.n++; return f.n }

func newSvcDB(t *testing.T) (Service, *gorm.DB) {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY, username TEXT UNIQUE, email TEXT UNIQUE, password_hash TEXT,
			display_name TEXT, avatar TEXT, role INTEGER DEFAULT 0, status INTEGER DEFAULT 0,
			group_id INTEGER, invite_code TEXT UNIQUE, invited_by INTEGER NOT NULL DEFAULT 0,
			email_verified_at DATETIME, last_login_at DATETIME, last_login_ip TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
		);
		CREATE TABLE user_groups (
			id INTEGER PRIMARY KEY, name TEXT UNIQUE, display_name TEXT, ratio REAL,
			priority INTEGER, status INTEGER, created_at DATETIME, updated_at DATETIME
		);
		INSERT INTO user_groups VALUES (1, 'default', '普通', 1.0, 0, 0, datetime('now'), datetime('now'));
	`).Error; err != nil {
		t.Fatal(err)
	}
	gsvc := group.NewService(group.NewRepository(db), clock.Real, &fakeIDGen{})
	return NewService(NewRepository(db), gsvc, &fakeIDGen{}, clock.Real), db
}

func TestUserService_CreateFillsGroupDefault(t *testing.T) {
	s, _ := newSvcDB(t)
	u, err := s.Create(context.Background(), CreateInput{Username: "alice", Email: ptr("a@example.com")})
	if err != nil {
		t.Fatal(err)
	}
	if u.GroupID == nil || *u.GroupID != 1 {
		t.Fatalf("default group not filled: %+v", u.GroupID)
	}
}

func TestUserService_CreateDuplicateEmail(t *testing.T) {
	s, _ := newSvcDB(t)
	_, _ = s.Create(context.Background(), CreateInput{Username: "alice", Email: ptr("a@example.com")})
	_, err := s.Create(context.Background(), CreateInput{Username: "bob", Email: ptr("a@example.com")})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeEmailRegistered {
		t.Fatalf("want CodeEmailRegistered, got %v", err)
	}
}

func TestUserService_CreateDuplicateUsername(t *testing.T) {
	s, _ := newSvcDB(t)
	_, _ = s.Create(context.Background(), CreateInput{Username: "alice"})
	_, err := s.Create(context.Background(), CreateInput{Username: "alice"})
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeUsernameTaken {
		t.Fatalf("want CodeUsernameTaken, got %v", err)
	}
}

func TestUserService_GetByEmail_NotFound(t *testing.T) {
	s, _ := newSvcDB(t)
	got, err := s.GetByEmail(context.Background(), "no@x.com")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("want nil")
	}
}

func TestUserService_Update_PatchOnlyNonNil(t *testing.T) {
	s, _ := newSvcDB(t)
	u, _ := s.Create(context.Background(), CreateInput{Username: "alice"})
	name := "Alice L"
	got, err := s.Update(context.Background(), u.ID, UpdateInput{DisplayName: &name})
	if err != nil {
		t.Fatal(err)
	}
	if got.DisplayName == nil || *got.DisplayName != "Alice L" {
		t.Fatalf("display_name not updated: %+v", got.DisplayName)
	}
}

func TestUserService_UpdatePasswordHash(t *testing.T) {
	s, _ := newSvcDB(t)
	u, _ := s.Create(context.Background(), CreateInput{Username: "alice"})
	if err := s.UpdatePasswordHash(context.Background(), u.ID, "newhash"); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetByID(context.Background(), u.ID)
	if got.PasswordHash == nil || *got.PasswordHash != "newhash" {
		t.Fatalf("hash not updated: %+v", got.PasswordHash)
	}
}

func TestUserService_MarkEmailVerified(t *testing.T) {
	s, _ := newSvcDB(t)
	u, _ := s.Create(context.Background(), CreateInput{Username: "alice", Status: StatusPendingEmailVerify})
	if err := s.MarkEmailVerified(context.Background(), u.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetByID(context.Background(), u.ID)
	if got.EmailVerifiedAt == nil {
		t.Fatal("email_verified_at not set")
	}
	if got.Status != StatusActive {
		t.Fatalf("want StatusActive, got %d", got.Status)
	}
}

func TestUserService_TouchLogin(t *testing.T) {
	s, _ := newSvcDB(t)
	u, _ := s.Create(context.Background(), CreateInput{Username: "alice"})
	_ = s.TouchLogin(context.Background(), u.ID, "9.9.9.9")
	got, _ := s.GetByID(context.Background(), u.ID)
	if got.LastLoginIP == nil || *got.LastLoginIP != "9.9.9.9" {
		t.Fatalf("ip not set: %+v", got.LastLoginIP)
	}
	if got.LastLoginAt == nil || time.Since(*got.LastLoginAt) > time.Second {
		t.Fatalf("last_login_at not updated: %+v", got.LastLoginAt)
	}
}

func TestUserService_Delete(t *testing.T) {
	s, _ := newSvcDB(t)
	u, _ := s.Create(context.Background(), CreateInput{Username: "alice"})
	if err := s.Delete(context.Background(), u.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetByID(context.Background(), u.ID)
	if got != nil {
		t.Fatalf("want soft deleted, got %+v", got)
	}
}
