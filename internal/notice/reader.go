package notice

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// Reader 用 Redis SET 维护用户已读集合。
type Reader interface {
	// MarkRead 把 noticeID 加入 user 的已读集合;幂等(SADD)。
	MarkRead(ctx context.Context, userID, noticeID int64) error
	// IsRead 单条已读判定。
	IsRead(ctx context.Context, userID, noticeID int64) (bool, error)
	// ReadSet 返回 user 所有已读 notice id。
	ReadSet(ctx context.Context, userID int64) (map[int64]struct{}, error)
	// UnreadCount 给一组候选 id,返回未读数量。
	UnreadCount(ctx context.Context, userID int64, visibleIDs []int64) (int, error)
}

type reader struct {
	rdb *redis.Client
}

// NewReader 构造 Reader 默认实现。
func NewReader(rdb *redis.Client) Reader { return &reader{rdb: rdb} }

func readKey(userID int64) string {
	return "notice:read:" + strconv.FormatInt(userID, 10)
}

// MarkRead 幂等 SADD。无 TTL(永久存储)。
func (r *reader) MarkRead(ctx context.Context, userID, noticeID int64) error {
	return r.rdb.SAdd(ctx, readKey(userID), noticeID).Err()
}

// IsRead 单条 SISMEMBER。
func (r *reader) IsRead(ctx context.Context, userID, noticeID int64) (bool, error) {
	return r.rdb.SIsMember(ctx, readKey(userID), noticeID).Result()
}

// ReadSet 返回 user 所有已读 id;空集返空 map。
func (r *reader) ReadSet(ctx context.Context, userID int64) (map[int64]struct{}, error) {
	members, err := r.rdb.SMembers(ctx, readKey(userID)).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[int64]struct{}, len(members))
	for _, s := range members {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			continue
		}
		out[n] = struct{}{}
	}
	return out, nil
}

// unreadBatchSize 单次 SMISMEMBER 的最大候选数。
const unreadBatchSize = 500

// UnreadCount 用 SMISMEMBER 分批查询,统计未读数。
//
// 大集合(>500) 拆批避免单次 redis command 过大。
func (r *reader) UnreadCount(ctx context.Context, userID int64, visibleIDs []int64) (int, error) {
	if len(visibleIDs) == 0 {
		return 0, nil
	}
	key := readKey(userID)
	unread := 0
	for start := 0; start < len(visibleIDs); start += unreadBatchSize {
		end := start + unreadBatchSize
		if end > len(visibleIDs) {
			end = len(visibleIDs)
		}
		batch := visibleIDs[start:end]
		args := make([]any, len(batch))
		for i, id := range batch {
			args[i] = id
		}
		res, err := r.rdb.SMIsMember(ctx, key, args...).Result()
		if err != nil {
			return 0, err
		}
		for _, isMember := range res {
			if !isMember {
				unread++
			}
		}
	}
	return unread, nil
}
