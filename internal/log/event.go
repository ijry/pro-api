package log

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

// Event 是请求/消费同表的一条日志（对应 request_logs）。
type Event struct {
	ID                  int64     `gorm:"column:id"                  json:"id,string"`
	CreatedAt           time.Time `gorm:"column:created_at"          json:"created_at"`
	UserID              int64     `gorm:"column:user_id"             json:"user_id,string"`
	TokenID             int64     `gorm:"column:token_id"            json:"token_id,string"`
	DeptID              *int64    `gorm:"column:dept_id"             json:"dept_id,omitempty,string"`
	GroupID             *int64    `gorm:"column:group_id"            json:"group_id,omitempty,string"`
	EventType           int8      `gorm:"column:event_type"          json:"event_type"`
	ClientModel         string    `gorm:"column:client_model"        json:"client_model"`
	UpstreamModel       string    `gorm:"column:upstream_model"      json:"upstream_model,omitempty"`
	ChannelID           *int64    `gorm:"column:channel_id"          json:"channel_id,omitempty,string"`
	Protocol            string    `gorm:"column:protocol"            json:"protocol"`
	Endpoint            string    `gorm:"column:endpoint"            json:"endpoint"`
	IP                  string    `gorm:"column:ip"                  json:"ip,omitempty"`
	UserAgent           string    `gorm:"column:user_agent"          json:"user_agent,omitempty"`
	StatusCode          int       `gorm:"column:status_code"         json:"status_code"`
	LatencyMS           int       `gorm:"column:latency_ms"          json:"latency_ms"`
	TTFTMS              int       `gorm:"column:ttft_ms"             json:"ttft_ms,omitempty"`
	Stream              bool      `gorm:"column:stream"              json:"stream"`
	InputTokens         int       `gorm:"column:input_tokens"        json:"input_tokens"`
	OutputTokens        int       `gorm:"column:output_tokens"       json:"output_tokens"`
	CachedTokens        int       `gorm:"column:cached_tokens"       json:"cached_tokens"`
	ReasoningTokens     int       `gorm:"column:reasoning_tokens"    json:"reasoning_tokens"`
	TotalQuota          int64     `gorm:"column:total_quota"         json:"total_quota,string"`
	BillingInputRatio   float64   `gorm:"column:billing_input_ratio"  json:"billing_input_ratio"`
	BillingOutputRatio  float64   `gorm:"column:billing_output_ratio" json:"billing_output_ratio"`
	BillingGroupRatio   float64   `gorm:"column:billing_group_ratio"  json:"billing_group_ratio"`
	ErrorCode           int       `gorm:"column:error_code"          json:"error_code,omitempty"`
	ErrorMsg            string    `gorm:"column:error_msg"           json:"error_msg,omitempty"`
	TraceID             string    `gorm:"column:trace_id"            json:"trace_id,omitempty"`
}

// TableName implements gorm.Tabler.
func (Event) TableName() string { return "request_logs" }

// ErrorEvent 对应 error_logs。
type ErrorEvent struct {
	ID        int64           `gorm:"column:id"          json:"id,string"`
	CreatedAt time.Time       `gorm:"column:created_at"  json:"created_at"`
	UserID    *int64          `gorm:"column:user_id"     json:"user_id,omitempty,string"`
	TokenID   *int64          `gorm:"column:token_id"    json:"token_id,omitempty,string"`
	ChannelID *int64          `gorm:"column:channel_id"  json:"channel_id,omitempty,string"`
	ErrorCode int             `gorm:"column:error_code"  json:"error_code"`
	ErrorType string          `gorm:"column:error_type"  json:"error_type"`
	Stack     string          `gorm:"column:stack"       json:"stack,omitempty"`
	Context   json.RawMessage `gorm:"column:context"     json:"context,omitempty"`
	TraceID   string          `gorm:"column:trace_id"    json:"trace_id,omitempty"`
}

// TableName implements gorm.Tabler.
func (ErrorEvent) TableName() string { return "error_logs" }

// AuditEntry 对应 audit_logs（只读，写入在各业务模块）。
type AuditEntry struct {
	ID         int64           `gorm:"column:id"          json:"id,string"`
	CreatedAt  time.Time       `gorm:"column:created_at"  json:"created_at"`
	ActorID    *int64          `gorm:"column:actor_id"    json:"actor_id,omitempty,string"`
	ActorRole  int8            `gorm:"column:actor_role"  json:"actor_role"`
	Action     string          `gorm:"column:action"      json:"action"`
	TargetType string          `gorm:"column:target_type" json:"target_type"`
	TargetID   *int64          `gorm:"column:target_id"   json:"target_id,omitempty,string"`
	Before     json.RawMessage `gorm:"column:before"      json:"before,omitempty"`
	After      json.RawMessage `gorm:"column:after"       json:"after,omitempty"`
	IP         string          `gorm:"column:ip"          json:"ip,omitempty"`
}

// TableName implements gorm.Tabler.
func (AuditEntry) TableName() string { return "audit_logs" }

// Cursor 是 keyset 翻页游标。
type Cursor struct {
	CreatedAt time.Time
	ID        int64
}

// String 把 cursor 编码为 base64url 字符串：base64url("RFC3339Nano|id")。
func (c Cursor) String() string {
	raw := c.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + strconv.FormatInt(c.ID, 10)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// ParseCursor 解码 base64url cursor。
func ParseCursor(s string) (*Cursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidCursor
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, ErrInvalidCursor
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	return &Cursor{CreatedAt: t.UTC(), ID: id}, nil
}

// ErrInvalidCursor 表示 cursor 解析失败。
var ErrInvalidCursor = errors.New("log: invalid cursor")

// Query 是 request_logs 查询请求。
type Query struct {
	UserID      *int64
	TokenID     *int64
	ChannelID   *int64
	ClientModel string  // "" = 不过滤；支持通配 "gpt-4*" → LIKE 'gpt-4%'
	EventType   *int8
	StatusCode  *int
	From        time.Time
	To          time.Time
	TraceID     string

	Cursor   *Cursor
	PageSize int
}

// QueryResult 查询结果。
type QueryResult struct {
	Items      []Event
	NextCursor *Cursor
	Total      *int64
}

// ErrorQuery 是 error_logs 查询请求。
type ErrorQuery struct {
	UserID    *int64
	TokenID   *int64
	ChannelID *int64
	ErrorCode *int
	ErrorType string
	From      time.Time
	To        time.Time
	TraceID   string

	Cursor   *Cursor
	PageSize int
}

// ErrorQueryResult 错误日志查询结果。
type ErrorQueryResult struct {
	Items      []ErrorEvent
	NextCursor *Cursor
	Total      *int64
}

// AuditQuery 查 audit_logs。
type AuditQuery struct {
	ActorID    *int64
	Action     string
	TargetType string
	TargetID   *int64
	From       time.Time
	To         time.Time

	Cursor   *Cursor
	PageSize int
}

// AuditQueryResult 审计日志查询结果。
type AuditQueryResult struct {
	Items      []AuditEntry
	NextCursor *Cursor
	Total      *int64
}
