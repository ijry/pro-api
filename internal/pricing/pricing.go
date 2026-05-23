package pricing

import (
	"context"
	"errors"
	"sync"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IDGenerator 是 pricing 模块对 idgen 的最小依赖。
type IDGenerator interface {
	Generate() int64
}

// ChannelInfo 是 M1-05 channel.Channel 的最小依赖。
// 通过本地接口解耦,允许 M1-05 在并行实施期间不阻塞 M1-06。
//
// 实施期 nil 容忍:Compute / RatioFor 在 channel == nil 时跳过 ModelOverride 步骤。
type ChannelInfo interface {
	// ModelOverrideFor 返回某模型在此渠道下的倍率覆盖,bool 表明是否有覆盖。
	ModelOverrideFor(model string) (Ratios, bool)
}

// GroupRatioLookup 让 wire 注入"按 user group id 取倍率"的查询(M1-05 group 包提供)。
// 实施期可 nil,nil 时返回默认 1.0(普通组)。
type GroupRatioLookup func(ctx context.Context, groupID int64) float64

// CatalogInfo 是模型字典的最小依赖(M1-04 model_catalogs)。
// 实施期可 nil,nil 时所有默认返回零值,启发式按 4096 兜底。
type CatalogInfo interface {
	LookupCatalog(model string) Catalog
}

// Catalog 是模型字典快照。0 值字段表示"未设置"。
type Catalog struct {
	Model                 string
	MaxInputTokens        int
	DefaultInputRatio     float64
	DefaultOutputRatio    float64
	DefaultCachedRatio    float64
	DefaultReasoningRatio float64
}

// Ratios 是计费倍率快照(给日志 / 对账写回)。
type Ratios struct {
	Input     float64
	Output    float64
	Cached    float64
	Reasoning float64
	Group     float64
}

// EstimateInput 是 EstimateMax 的输入。
type EstimateInput struct {
	InputTokens       int
	MaxOutTokens      int
	Stream            bool
	BillingGroupRatio float64
}

// ComputeInput 是 Compute 的输入。
type ComputeInput struct {
	Model           string
	GroupID         int64
	Channel         ChannelInfo // 可 nil(估算路径)
	InputTokens     int
	OutputTokens    int
	CachedTokens    int
	ReasoningTokens int
}

// ComputeResult 是 Compute 的输出。
type ComputeResult struct {
	Quota  int64
	Ratios Ratios
}

// Pricing 是计费倍率的核心接口。
type Pricing interface {
	// Compute 计算实际 quota 消耗。
	Compute(ctx context.Context, in ComputeInput) ComputeResult

	// RatioFor 返回某模型/分组/渠道生效的倍率(给 Reserve 估算前的预算评估)。
	RatioFor(ctx context.Context, model string, groupID int64, ch ChannelInfo) Ratios

	// EstimateMax 用 max_tokens 估算最大可能消耗(用于 Reserve)。
	EstimateMax(ctx context.Context, model string, in EstimateInput) int64

	// DefaultMaxOut 返回模型默认 max_out_tokens(无 max_tokens 时的 fallback)。
	DefaultMaxOut(ctx context.Context, model string) int
}

// Service 是 Pricing + 规则 CRUD,给管理员 handler 用。
type Service interface {
	Pricing

	// List 列规则(分页 + 过滤)。
	List(ctx context.Context, filter ListFilter) ([]*Rule, int64, error)
	// Get 按 id 取。
	Get(ctx context.Context, id int64) (*Rule, error)
	// Create 新建。
	Create(ctx context.Context, in CreateInput, actor int64) (*Rule, error)
	// Update 部分字段更新(支持 clear_* 把字段置 NULL)。
	Update(ctx context.Context, id int64, patch UpdatePatch, actor int64) (*Rule, error)
	// Delete 物理删除。
	Delete(ctx context.Context, id int64, actor int64) error

	// Refresh 强制从 DB 重载缓存(用于测试 / 管理员触发)。
	Refresh(ctx context.Context) error

	// Close 关停 Pub/Sub 订阅。
	Close() error
}

// CreateInput 是 Create 的入参。
type CreateInput struct {
	Scope          string
	GroupID        *int64
	Model          *string
	InputRatio     *float64
	OutputRatio    *float64
	CachedRatio    *float64
	ReasoningRatio *float64
	Priority       int16
	Status         int8
}

// UpdatePatch 是 Update 的入参。
type UpdatePatch struct {
	InputRatio     *float64
	OutputRatio    *float64
	CachedRatio    *float64
	ReasoningRatio *float64
	Priority       *int16
	Status         *int8

	// clear_* 让管理员明确"把这个倍率重置为 NULL"。
	ClearInput     bool
	ClearOutput    bool
	ClearCached    bool
	ClearReasoning bool
}

// ListFilter 是 List 的过滤参数。
type ListFilter struct {
	Scope   string
	GroupID *int64
	Model   *string
	Status  *int8
	Page    int
	Size    int
}

// Config 是 New 的参数。
type Config struct {
	DB           *gorm.DB
	Cache        *redis.Client
	Log          *zap.Logger
	Clock        clock.Clock
	IDGen        IDGenerator
	Audit        audit.Logger
	GroupRatio   GroupRatioLookup // 可 nil,按 1.0 兜底
	Catalog      CatalogInfo      // 可 nil,按零值兜底
	DefaultMaxOut int             // 兜底,0 时用 4096

	// 缓存 / Pub/Sub 配置
	InvalidateChannel string // 默认 "proapi:pricing:invalidate"
}

// PubSubChannel 是 pricing 规则失效广播 channel 名。
const PubSubChannel = "proapi:pricing:invalidate"

// New 构造一个 Service。
//
// 内部会:
//  1. 立即从 DB SELECT 全部 enabled 规则到 sync.Map 缓存
//  2. 订阅 PubSubChannel 收到失效广播后重载
//
// 若 Cache 为 nil,Pub/Sub 步骤跳过。
func New(ctx context.Context, cfg Config) (Service, error) {
	if cfg.DB == nil {
		return nil, errors.New("pricing: Config.DB is nil")
	}
	if cfg.IDGen == nil {
		return nil, errors.New("pricing: Config.IDGen is nil")
	}
	if cfg.Log == nil {
		cfg.Log = zap.NewNop()
	}
	if cfg.Clock == nil {
		cfg.Clock = clock.Real
	}
	if cfg.Audit == nil {
		cfg.Audit = audit.NewNoop()
	}
	if cfg.InvalidateChannel == "" {
		cfg.InvalidateChannel = PubSubChannel
	}
	if cfg.DefaultMaxOut == 0 {
		cfg.DefaultMaxOut = 4096
	}
	s := &service{
		db:                cfg.DB,
		rdb:               cfg.Cache,
		log:               cfg.Log,
		clk:               cfg.Clock,
		idgen:             cfg.IDGen,
		audit:             cfg.Audit,
		groupRatio:        cfg.GroupRatio,
		catalog:           cfg.Catalog,
		defaultMaxOut:     cfg.DefaultMaxOut,
		invalidateChannel: cfg.InvalidateChannel,
		cache:             newRuleCache(),
		stopCh:            make(chan struct{}),
	}
	if err := s.Refresh(ctx); err != nil {
		return nil, err
	}
	if cfg.Cache != nil {
		s.sub = cfg.Cache.Subscribe(ctx, cfg.InvalidateChannel)
		s.wg.Add(1)
		go s.runInvalidator()
	}
	return s, nil
}

// service 是 Service 的默认实现。
type service struct {
	db                *gorm.DB
	rdb               *redis.Client
	log               *zap.Logger
	clk               clock.Clock
	idgen             IDGenerator
	audit             audit.Logger
	groupRatio        GroupRatioLookup
	catalog           CatalogInfo
	defaultMaxOut     int
	invalidateChannel string

	cache *ruleCache

	sub    *redis.PubSub
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// Close 关停 Pub/Sub 订阅。
func (s *service) Close() error {
	if s.stopCh != nil {
		select {
		case <-s.stopCh:
		default:
			close(s.stopCh)
		}
	}
	if s.sub != nil {
		_ = s.sub.Close()
	}
	s.wg.Wait()
	return nil
}

func (s *service) runInvalidator() {
	defer s.wg.Done()
	ch := s.sub.Channel()
	for {
		select {
		case <-s.stopCh:
			return
		case _, ok := <-ch:
			if !ok {
				return
			}
			if err := s.Refresh(context.Background()); err != nil {
				s.log.Warn("pricing: refresh after invalidate failed", zap.Error(err))
			}
		}
	}
}

// 占位:把 invalidateChannel 推一条消息到 Redis Pub/Sub。
func (s *service) publishInvalidate(ctx context.Context) {
	if s.rdb == nil {
		return
	}
	if err := s.rdb.Publish(ctx, s.invalidateChannel, "*").Err(); err != nil {
		s.log.Warn("pricing: publish invalidate failed", zap.Error(err))
	}
}
