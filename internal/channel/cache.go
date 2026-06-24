package channel

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	pubSubInvalidate = "proapi:channel:invalidate"
)

// channelCache 是内存本地缓存。
type channelCache struct {
	repo *repo
	rdb  *redis.Client
	log  *zap.Logger

	mu               sync.RWMutex
	byID             map[int64]*Channel
	channelsByModel  map[string][]*Channel
	decryptFailedIDs map[int64]struct{}

	pubsub *redis.PubSub
	stop   chan struct{}
}

func newChannelCache(r *repo, rdb *redis.Client, log *zap.Logger) *channelCache {
	return &channelCache{
		repo:             r,
		rdb:              rdb,
		log:              log,
		byID:             make(map[int64]*Channel),
		channelsByModel:  make(map[string][]*Channel),
		decryptFailedIDs: make(map[int64]struct{}),
		stop:             make(chan struct{}),
	}
}

// Warmup 全量加载。返回 loaded, decryptFailed, err。
func (c *channelCache) Warmup(ctx context.Context) (loaded int, decryptFailed []int64, err error) {
	channels, err := c.repo.ListAllActive(ctx)
	if err != nil {
		return 0, nil, err
	}

	ids := int64SliceIDs(channels)
	mappings, err := c.repo.ListMappingsByChannelIDs(ctx, ids)
	if err != nil {
		return 0, nil, err
	}

	// 按 channel_id 分组
	byChannel := make(map[int64][]*ModelMapping, len(channels))
	for _, m := range mappings {
		byChannel[m.ChannelID] = append(byChannel[m.ChannelID], m)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 清空旧索引
	c.byID = make(map[int64]*Channel)
	c.channelsByModel = make(map[string][]*Channel)
	c.decryptFailedIDs = make(map[int64]struct{})

	for _, ch := range channels {
		if err2 := c.repo.hydrate(ch); err2 != nil {
			c.log.Warn("cache: decrypt failed, skipping channel",
				zap.Int64("id", ch.ID), zap.Error(err2))
			c.decryptFailedIDs[ch.ID] = struct{}{}
			decryptFailed = append(decryptFailed, ch.ID)
			continue
		}
		ch.ModelMap, ch.ModelOverrides = buildModelMap(byChannel[ch.ID])
		c.byID[ch.ID] = ch
		for clientModel := range ch.ModelMap {
			c.channelsByModel[clientModel] = append(c.channelsByModel[clientModel], ch)
		}
		loaded++
	}

	// 对每个模型列表按 priority DESC, weight DESC 排序
	for model := range c.channelsByModel {
		list := c.channelsByModel[model]
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].Priority != list[j].Priority {
				return list[i].Priority > list[j].Priority
			}
			return list[i].Weight > list[j].Weight
		})
	}
	return loaded, decryptFailed, nil
}

// StartPubSub 启动 Pub/Sub 订阅循环。
func (c *channelCache) StartPubSub(ctx context.Context) {
	c.pubsub = c.rdb.Subscribe(ctx, pubSubInvalidate)
	go c.loop(ctx)
}

func (c *channelCache) loop(ctx context.Context) {
	ch := c.pubsub.Channel()
	for {
		select {
		case <-c.stop:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			id, err := strconv.ParseInt(msg.Payload, 10, 64)
			if err != nil {
				c.log.Warn("cache: invalid pubsub payload", zap.String("payload", msg.Payload))
				continue
			}
			if err := c.Reload(ctx, id); err != nil {
				c.log.Error("cache: reload failed", zap.Int64("id", id), zap.Error(err))
			}
		}
	}
}

// Reload 单条更新。
func (c *channelCache) Reload(ctx context.Context, id int64) error {
	ch, err := c.repo.GetByIDIncludingDeleted(ctx, id)
	if err != nil {
		return err
	}

	mappings, _ := c.repo.ListMappings(ctx, id)

	c.mu.Lock()
	defer c.mu.Unlock()

	// 先从旧索引移除
	old, ok := c.byID[id]
	if ok {
		for clientModel := range old.ModelMap {
			c.channelsByModel[clientModel] = removeChannelFromSlice(c.channelsByModel[clientModel], id)
			if len(c.channelsByModel[clientModel]) == 0 {
				delete(c.channelsByModel, clientModel)
			}
		}
		delete(c.byID, id)
	}

	// 如果已软删或不存在，直接结束
	if ch == nil || ch.DeletedAt != nil {
		return nil
	}

	// 重新 hydrate
	if err2 := c.repo.hydrate(ch); err2 != nil {
		c.log.Warn("cache: reload hydrate failed", zap.Int64("id", id), zap.Error(err2))
		return nil
	}
	ch.ModelMap, ch.ModelOverrides = buildModelMap(mappings)

	c.byID[id] = ch
	for clientModel := range ch.ModelMap {
		list := c.channelsByModel[clientModel]
		list = append(list, ch)
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].Priority != list[j].Priority {
				return list[i].Priority > list[j].Priority
			}
			return list[i].Weight > list[j].Weight
		})
		c.channelsByModel[clientModel] = list
	}
	return nil
}

// ChannelsByModel 返回某模型候选 channel 的深拷贝列表。
func (c *channelCache) ChannelsByModel(model string) []*Channel {
	c.mu.RLock()
	list := c.channelsByModel[model]
	c.mu.RUnlock()

	out := make([]*Channel, len(list))
	for i, ch := range list {
		cp := *ch
		cp.ModelMap = cloneStringMap(ch.ModelMap)
		cp.ModelOverrides = cloneOverrideMap(ch.ModelOverrides)
		cp.Cred = ch.Cred
		cp.Cred.Extra = cloneStringMap(ch.Cred.Extra)
		out[i] = &cp
	}
	return out
}

// ActiveModels 返回当前可调用模型。groupID > 0 时只返回全局渠道(group_id=0)
// 或同组渠道支持的模型。
func (c *channelCache) ActiveModels(groupID int64) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	seen := make(map[string]struct{}, len(c.channelsByModel))
	for model, list := range c.channelsByModel {
		for _, ch := range list {
			if ch.Status != 0 {
				continue
			}
			if groupID > 0 && ch.GroupID != 0 && ch.GroupID != groupID {
				continue
			}
			seen[model] = struct{}{}
			break
		}
	}
	models := make([]string, 0, len(seen))
	for model := range seen {
		models = append(models, model)
	}
	sort.Strings(models)
	return models
}

// GetByID 从缓存取一条(已解密)。
func (c *channelCache) GetByID(id int64) (*Channel, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ch, ok := c.byID[id]
	if !ok {
		return nil, false
	}
	cp := *ch
	cp.ModelMap = cloneStringMap(ch.ModelMap)
	cp.ModelOverrides = cloneOverrideMap(ch.ModelOverrides)
	cp.Cred = ch.Cred
	cp.Cred.Extra = cloneStringMap(ch.Cred.Extra)
	return &cp, true
}

// All 返回所有 channel 的深拷贝。
func (c *channelCache) All() []*Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*Channel, 0, len(c.byID))
	for _, ch := range c.byID {
		cp := *ch
		cp.ModelMap = cloneStringMap(ch.ModelMap)
		out = append(out, &cp)
	}
	return out
}

// Close 停止 Pub/Sub 循环。
func (c *channelCache) Close() error {
	close(c.stop)
	if c.pubsub != nil {
		return c.pubsub.Close()
	}
	return nil
}

// Publish 发送失效通知。
func (c *channelCache) Publish(ctx context.Context, channelID int64) {
	c.rdb.Publish(ctx, pubSubInvalidate, strconv.FormatInt(channelID, 10))
}

// helper functions

func removeChannelFromSlice(list []*Channel, id int64) []*Channel {
	out := list[:0]
	for _, ch := range list {
		if ch.ID != id {
			out = append(out, ch)
		}
	}
	return out
}

func cloneStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func cloneOverrideMap(m map[string]ModelOverride) map[string]ModelOverride {
	if m == nil {
		return nil
	}
	cp := make(map[string]ModelOverride, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

// tagsMatch 判断 channel 的 tags 是否包含 required 中所有 tag。
func tagsMatch(channelTagsJSON json.RawMessage, required []string) bool {
	if len(required) == 0 {
		return true
	}
	var tags []string
	_ = json.Unmarshal(channelTagsJSON, &tags)
	for _, req := range required {
		found := false
		for _, t := range tags {
			if t == req {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// inInt64Slice 判断 id 是否在切片中。
func inInt64Slice(slice []int64, id int64) bool {
	for _, v := range slice {
		if v == id {
			return true
		}
	}
	return false
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
