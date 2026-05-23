package log

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *dbStore) Query(ctx context.Context, q Query) (QueryResult, error) {
	if q.From.IsZero() || q.To.IsZero() {
		return QueryResult{}, ErrTimeRangeRequired
	}
	if q.To.Sub(q.From) > maxQueryRangeDays*24*time.Hour {
		return QueryResult{}, ErrTimeRangeTooWide
	}
	if q.PageSize <= 0 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}

	tx := s.db.WithContext(ctx).Model(&Event{}).
		Where("created_at >= ? AND created_at < ?", q.From.UTC(), q.To.UTC())

	if q.UserID != nil {
		tx = tx.Where("user_id = ?", *q.UserID)
	}
	if q.TokenID != nil {
		tx = tx.Where("token_id = ?", *q.TokenID)
	}
	if q.ChannelID != nil {
		tx = tx.Where("channel_id = ?", *q.ChannelID)
	}
	if q.EventType != nil {
		tx = tx.Where("event_type = ?", *q.EventType)
	}
	if q.StatusCode != nil {
		tx = tx.Where("status_code = ?", *q.StatusCode)
	}
	if q.TraceID != "" {
		tx = tx.Where("trace_id = ?", q.TraceID)
	}
	if q.ClientModel != "" {
		if strings.ContainsRune(q.ClientModel, '*') {
			pat := strings.ReplaceAll(q.ClientModel, "*", "%")
			tx = tx.Where("client_model LIKE ?", pat)
		} else {
			tx = tx.Where("client_model = ?", q.ClientModel)
		}
	}
	if q.Cursor != nil {
		tx = tx.Where("(created_at, id) < (?, ?)", q.Cursor.CreatedAt.UTC(), q.Cursor.ID)
	}

	var items []Event
	err := tx.Order("created_at DESC, id DESC").
		Limit(q.PageSize + 1).
		Find(&items).Error
	if err != nil {
		return QueryResult{}, err
	}

	res := QueryResult{Items: items}
	if len(items) > q.PageSize {
		res.Items = items[:q.PageSize]
		last := items[q.PageSize-1]
		res.NextCursor = &Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}

	// Total count only for small first pages
	if q.PageSize <= 50 && q.Cursor == nil {
		var total int64
		if err2 := tx.Session(&gorm.Session{}).Limit(-1).Offset(-1).Count(&total).Error; err2 == nil {
			res.Total = &total
		}
	}
	return res, nil
}

func (s *dbStore) QueryErrors(ctx context.Context, q ErrorQuery) (ErrorQueryResult, error) {
	if q.From.IsZero() || q.To.IsZero() {
		return ErrorQueryResult{}, ErrTimeRangeRequired
	}
	if q.PageSize <= 0 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}

	tx := s.db.WithContext(ctx).Model(&ErrorEvent{}).
		Where("created_at >= ? AND created_at < ?", q.From.UTC(), q.To.UTC())

	if q.UserID != nil {
		tx = tx.Where("user_id = ?", *q.UserID)
	}
	if q.TokenID != nil {
		tx = tx.Where("token_id = ?", *q.TokenID)
	}
	if q.ChannelID != nil {
		tx = tx.Where("channel_id = ?", *q.ChannelID)
	}
	if q.ErrorCode != nil {
		tx = tx.Where("error_code = ?", *q.ErrorCode)
	}
	if q.ErrorType != "" {
		tx = tx.Where("error_type = ?", q.ErrorType)
	}
	if q.TraceID != "" {
		tx = tx.Where("trace_id = ?", q.TraceID)
	}
	if q.Cursor != nil {
		tx = tx.Where("(created_at, id) < (?, ?)", q.Cursor.CreatedAt.UTC(), q.Cursor.ID)
	}

	var items []ErrorEvent
	err := tx.Order("created_at DESC, id DESC").
		Limit(q.PageSize + 1).
		Find(&items).Error
	if err != nil {
		return ErrorQueryResult{}, err
	}

	res := ErrorQueryResult{Items: items}
	if len(items) > q.PageSize {
		res.Items = items[:q.PageSize]
		last := items[q.PageSize-1]
		res.NextCursor = &Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return res, nil
}

func (s *dbStore) QueryAudits(ctx context.Context, q AuditQuery) (AuditQueryResult, error) {
	if q.From.IsZero() || q.To.IsZero() {
		return AuditQueryResult{}, ErrTimeRangeRequired
	}
	if q.PageSize <= 0 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}

	tx := s.db.WithContext(ctx).Model(&AuditEntry{}).
		Where("created_at >= ? AND created_at < ?", q.From.UTC(), q.To.UTC())

	if q.ActorID != nil {
		tx = tx.Where("actor_id = ?", *q.ActorID)
	}
	if q.TargetType != "" {
		tx = tx.Where("target_type = ?", q.TargetType)
	}
	if q.TargetID != nil {
		tx = tx.Where("target_id = ?", *q.TargetID)
	}
	if q.Action != "" {
		if strings.HasSuffix(q.Action, "*") {
			pat := strings.TrimSuffix(q.Action, "*") + "%"
			tx = tx.Where("action LIKE ?", pat)
		} else {
			tx = tx.Where("action = ?", q.Action)
		}
	}
	if q.Cursor != nil {
		tx = tx.Where("(created_at, id) < (?, ?)", q.Cursor.CreatedAt.UTC(), q.Cursor.ID)
	}

	var items []AuditEntry
	err := tx.Order("created_at DESC, id DESC").
		Limit(q.PageSize + 1).
		Find(&items).Error
	if err != nil {
		return AuditQueryResult{}, err
	}

	res := AuditQueryResult{Items: items}
	if len(items) > q.PageSize {
		res.Items = items[:q.PageSize]
		last := items[q.PageSize-1]
		res.NextCursor = &Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return res, nil
}

// Ensure unused import used.
var _ = errors.New
