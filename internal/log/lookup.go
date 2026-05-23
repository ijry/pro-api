package log

import "context"

// ChannelNameLooker is a minimal interface to look up channel names.
// Implemented by the channel service to avoid circular deps.
type ChannelNameLooker interface {
	NamesByIDs(ctx context.Context, ids []int64) map[int64]string
}

// UserNameLooker is a minimal interface to look up usernames.
type UserNameLooker interface {
	UsernamesByIDs(ctx context.Context, ids []int64) map[int64]string
}

// NoopChannelLooker is a stub for testing.
type NoopChannelLooker struct{}

func (NoopChannelLooker) NamesByIDs(context.Context, []int64) map[int64]string { return nil }

// NoopUserLooker is a stub for testing.
type NoopUserLooker struct{}

func (NoopUserLooker) UsernamesByIDs(context.Context, []int64) map[int64]string { return nil }
