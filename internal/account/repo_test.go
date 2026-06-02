//go:build integration

package account_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/crypto"
	"github.com/ijry/pro-api/internal/util/idgen"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// newTestDB spins up a dockertest MySQL container and runs the account migrations.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Skipf("docker not available: %v", err)
	}
	if err := pool.Client.Ping(); err != nil {
		t.Skipf("docker daemon not reachable: %v", err)
	}
	res, err := pool.Run("mysql", "8.0", []string{
		"MYSQL_ROOT_PASSWORD=proapi",
		"MYSQL_DATABASE=proapi",
	})
	if err != nil {
		t.Fatalf("could not start mysql: %v", err)
	}
	t.Cleanup(func() { _ = pool.Purge(res) })

	dsn := fmt.Sprintf("root:proapi@tcp(127.0.0.1:%s)/proapi?charset=utf8mb4&parseTime=True&loc=UTC",
		res.GetPort("3306/tcp"))

	var db *gorm.DB
	pool.MaxWait = 90 * time.Second
	if err := pool.Retry(func() error {
		var openErr error
		db, openErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if openErr != nil {
			return openErr
		}
		sqlDB, _ := db.DB()
		return sqlDB.Ping()
	}); err != nil {
		t.Fatalf("could not connect to mysql: %v", err)
	}

	// Run account migrations via raw SQL.
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id                       BIGINT       NOT NULL PRIMARY KEY,
			channel_id               BIGINT       NOT NULL,
			share_tag                VARCHAR(64)  NULL,
			name                     VARCHAR(128) NOT NULL DEFAULT '',
			provider                 VARCHAR(32)  NOT NULL,
			tier                     VARCHAR(32)  NOT NULL DEFAULT 'unknown',
			cred_type                VARCHAR(16)  NOT NULL,
			email                    VARCHAR(128) NULL,
			external_account_id      VARCHAR(64)  NULL,
			credentials              TEXT         NOT NULL,
			priority                 SMALLINT     NOT NULL DEFAULT 0,
			weight                   INT          NOT NULL DEFAULT 100,
			status                   TINYINT      NOT NULL DEFAULT 0,
			cooldown_until           DATETIME(3)  NULL,
			last_failure_at          DATETIME(3)  NULL,
			last_failure_reason      VARCHAR(256) NOT NULL DEFAULT '',
			consec_failures          INT          NOT NULL DEFAULT 0,
			last_success_at          DATETIME(3)  NULL,
			last_used_at             DATETIME(3)  NULL,
			quota_5h_total           BIGINT       NULL,
			quota_5h_remaining       BIGINT       NULL,
			quota_5h_reset_at        DATETIME(3)  NULL,
			quota_week_total         BIGINT       NULL,
			quota_week_remaining     BIGINT       NULL,
			quota_week_reset_at      DATETIME(3)  NULL,
			quota_synced_at          DATETIME(3)  NULL,
			access_token_expires_at  DATETIME(3)  NULL,
			refresh_token_valid      TINYINT      NOT NULL DEFAULT 0,
			last_refreshed_at        DATETIME(3)  NULL,
			import_source            VARCHAR(32)  NOT NULL DEFAULT '',
			extra                    JSON         NULL,
			created_at               DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			updated_at               DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
			deleted_at               DATETIME(3)  NULL,
			INDEX idx_accounts_channel_status (channel_id, status),
			INDEX idx_accounts_share_status (share_tag, status),
			INDEX idx_accounts_status_cooldown (status, cooldown_until),
			INDEX idx_accounts_provider_tier (provider, tier),
			INDEX idx_accounts_token_exp (access_token_expires_at, status),
			INDEX idx_accounts_deleted (deleted_at),
			UNIQUE KEY uk_accounts_provider_extid (provider, external_account_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS account_events (
			id          BIGINT       NOT NULL PRIMARY KEY,
			account_id  BIGINT       NOT NULL,
			event_type  VARCHAR(32)  NOT NULL,
			payload     JSON         NULL,
			created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
			INDEX idx_account_events_acc_created (account_id, created_at),
			INDEX idx_account_events_type_created (event_type, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, s := range sqls {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("migration failed: %v", err)
		}
	}
	return db
}

func TestRepo_CreateGet(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	cr, err := crypto.NewAESGCM([]byte("01234567890123456789012345678901"))
	require.NoError(t, err)
	idg, err := idgen.New(1)
	require.NoError(t, err)
	r := account.NewRepository(db, cr, idg, clock.Real)

	a := &account.Account{
		ChannelID: 100,
		Provider:  "anthropic",
		Tier:      "pro_max",
		CredType:  "oauth",
		Email:     "u@example.com",
		Status:    account.StatusActive,
		Priority:  0,
		Weight:    100,
		Extra:     json.RawMessage("{}"),
		Cred: account.AccountCred{
			AccessToken:  "at-1",
			RefreshToken: "rt-1",
			ExpiresAt:    time.Now().Add(time.Hour),
		},
	}
	require.NoError(t, r.Create(ctx, a))
	require.NotZero(t, a.ID)

	got, err := r.Get(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, "anthropic", got.Provider)
	require.Equal(t, "at-1", got.Cred.AccessToken, "credentials should be decrypted on Get")
}

func TestRepo_ListByChannel(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	cr, err := crypto.NewAESGCM([]byte("01234567890123456789012345678901"))
	require.NoError(t, err)
	idg, err := idgen.New(2)
	require.NoError(t, err)
	r := account.NewRepository(db, cr, idg, clock.Real)
	for i := 0; i < 3; i++ {
		require.NoError(t, r.Create(ctx, &account.Account{
			ChannelID: 200, Provider: "anthropic", CredType: "apikey",
			Status: account.StatusActive, Weight: 100,
			Extra: json.RawMessage("{}"),
			Cred:  account.AccountCred{APIKey: "sk-x"},
		}))
	}
	list, err := r.ListByChannel(ctx, 200)
	require.NoError(t, err)
	require.Len(t, list, 3)
}

func TestRepo_AppendEvent(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	cr, err := crypto.NewAESGCM([]byte("01234567890123456789012345678901"))
	require.NoError(t, err)
	idg, err := idgen.New(3)
	require.NoError(t, err)
	r := account.NewRepository(db, cr, idg, clock.Real)
	a := &account.Account{
		ChannelID: 300, Provider: "openai", CredType: "oauth",
		Status: account.StatusActive, Weight: 100,
		Extra: json.RawMessage("{}"),
		Cred:  account.AccountCred{AccessToken: "at"},
	}
	require.NoError(t, r.Create(ctx, a))
	require.NoError(t, r.AppendEvent(ctx, a.ID, "imported", map[string]any{"src": "test"}))
}
