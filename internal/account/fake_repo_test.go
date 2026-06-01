package account_test

import (
	"context"
	"sync"
	"time"

	"github.com/ijry/pro-api/internal/account"
)

type fakeRepoEvent struct {
	accountID int64
	eventType string
	payload   any
}

type fakeRepo struct {
	mu     sync.Mutex
	nextID int64
	items  map[int64]*account.Account
	events []fakeRepoEvent
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{items: map[int64]*account.Account{}}
}

func (f *fakeRepo) Create(_ context.Context, a *account.Account) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	if a.ID == 0 {
		a.ID = f.nextID
	}
	f.items[a.ID] = a
	return nil
}

func (f *fakeRepo) Update(_ context.Context, a *account.Account) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.items[a.ID] = a
	return nil
}

func (f *fakeRepo) Get(_ context.Context, id int64) (*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a := f.items[id]
	if a == nil {
		return nil, account.ErrNotFound
	}
	return a, nil
}

func (f *fakeRepo) ListByChannel(_ context.Context, channelID int64) ([]*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []*account.Account{}
	for _, a := range f.items {
		if a.ChannelID == channelID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (f *fakeRepo) ListByShareTag(_ context.Context, tag string) ([]*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []*account.Account{}
	for _, a := range f.items {
		if a.ShareTag == tag {
			out = append(out, a)
		}
	}
	return out, nil
}

func (f *fakeRepo) ListForRefresher(context.Context, time.Time, int) ([]*account.Account, error) {
	return nil, nil
}

func (f *fakeRepo) ListForReaper(_ context.Context, now time.Time, _ int) ([]*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []*account.Account{}
	for _, a := range f.items {
		if a.Status == account.StatusCooldown && a.CooldownUntil != nil && !a.CooldownUntil.After(now) {
			out = append(out, a)
		}
	}
	return out, nil
}

func (f *fakeRepo) Delete(_ context.Context, id int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.items, id)
	return nil
}

func (f *fakeRepo) AppendEvent(_ context.Context, id int64, t string, p any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, fakeRepoEvent{id, t, p})
	return nil
}
