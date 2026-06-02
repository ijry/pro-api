package invite

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
		CREATE TABLE invite_records (
			id INTEGER PRIMARY KEY,
			inviter_id INTEGER NOT NULL,
			invitee_id INTEGER NOT NULL,
			order_id   INTEGER NOT NULL,
			rebate_cents   INTEGER NOT NULL DEFAULT 0,
			rebate_credits INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME
		);
	`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCountByInviter(t *testing.T) {
	db := openTestDB(t)
	r := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	for i := 0; i < 3; i++ {
		db.Exec(`INSERT INTO invite_records (id,inviter_id,invitee_id,order_id,rebate_cents,rebate_credits,created_at) VALUES (?,100,?,999,10,10,?)`,
			int64(i+1), int64(200+i), now)
	}
	// unrelated inviter
	db.Exec(`INSERT INTO invite_records (id,inviter_id,invitee_id,order_id,rebate_cents,rebate_credits,created_at) VALUES (99,999,1,1,0,0,?)`, now)

	n, err := r.CountByInviter(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("want 3, got %d", n)
	}

	n0, _ := r.CountByInviter(ctx, 777)
	if n0 != 0 {
		t.Fatalf("want 0 for unknown inviter, got %d", n0)
	}
}
