package log

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// OverviewQuery — stats overview query parameters.
type OverviewQuery struct {
	From    time.Time
	To      time.Time
	UserID  *int64
	GroupID *int64
}

// Overview — aggregated stats for a time window.
type Overview struct {
	RequestCount int64   `json:"request_count"`
	UsageCount   int64   `json:"usage_count"`
	ErrorCount   int64   `json:"error_count"`
	ErrorRate    float64 `json:"error_rate"`
	TotalQuota   int64   `json:"total_quota"`
	ActiveUsers  int64   `json:"active_users"`
	ActiveTokens int64   `json:"active_tokens"`
	AvgLatencyMS float64 `json:"avg_latency_ms"`
	P95LatencyMS float64 `json:"p95_latency_ms"`
}

// TimeseriesQuery — timeseries query parameters.
type TimeseriesQuery struct {
	From        time.Time
	To          time.Time
	Granularity string // "hour" | "day"
	UserID      *int64
	ChannelID   *int64
	Model       string
}

// TimeseriesPoint — a single time bucket.
type TimeseriesPoint struct {
	Bucket       time.Time `json:"bucket"`
	RequestCount int64     `json:"request_count"`
	ErrorCount   int64     `json:"error_count"`
	TotalQuota   int64     `json:"total_quota"`
	InputTokens  int64     `json:"input_tokens"`
	OutputTokens int64     `json:"output_tokens"`
	AvgLatencyMS float64   `json:"avg_latency_ms"`
}

// Timeseries — collection of time buckets.
type Timeseries struct {
	Granularity string            `json:"granularity"`
	Points      []TimeseriesPoint `json:"points"`
}

// GroupQuery — group-by stats parameters.
type GroupQuery struct {
	From    time.Time
	To      time.Time
	UserID  *int64
	Limit   int
	OrderBy string // "quota" | "requests" | "tokens"
}

// GroupRow — a single group result row.
type GroupRow struct {
	Key          string  `json:"key"`
	Label        string  `json:"label"`
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`
	TotalQuota   int64   `json:"total_quota"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	AvgLatencyMS float64 `json:"avg_latency_ms"`
}

// GroupResult — group-by result with total count.
type GroupResult struct {
	Rows  []GroupRow `json:"rows"`
	Total int64      `json:"total"`
}

func (s *dbStore) StatsOverview(ctx context.Context, q OverviewQuery) (Overview, error) {
	if q.From.IsZero() || q.To.IsZero() {
		return Overview{}, ErrTimeRangeRequired
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var out Overview
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var row struct {
			ReqCnt int64
			ErrCnt int64
			Quota  int64
			AvgLat float64
		}
		err := s.baseQuery(egCtx, q).
			Select(`COUNT(*) AS req_cnt,
				SUM(CASE WHEN status_code >= 400 OR error_code != 0 THEN 1 ELSE 0 END) AS err_cnt,
				COALESCE(SUM(total_quota), 0) AS quota,
				COALESCE(AVG(latency_ms), 0) AS avg_lat`).
			Scan(&row).Error
		if err != nil {
			return err
		}
		out.RequestCount = row.ReqCnt
		out.ErrorCount = row.ErrCnt
		out.TotalQuota = row.Quota
		out.AvgLatencyMS = row.AvgLat
		if row.ReqCnt > 0 {
			out.ErrorRate = float64(row.ErrCnt) / float64(row.ReqCnt)
		}
		return nil
	})

	eg.Go(func() error {
		var n int64
		err := s.baseQuery(egCtx, q).
			Where("event_type = 1 AND status_code < 400").
			Count(&n).Error
		if err != nil {
			return err
		}
		out.UsageCount = n
		return nil
	})

	eg.Go(func() error {
		var n int64
		err := s.baseQuery(egCtx, q).
			Distinct("user_id").
			Count(&n).Error
		if err != nil {
			return err
		}
		out.ActiveUsers = n
		return nil
	})

	eg.Go(func() error {
		var n int64
		err := s.baseQuery(egCtx, q).
			Distinct("token_id").
			Count(&n).Error
		if err != nil {
			return err
		}
		out.ActiveTokens = n
		return nil
	})

	eg.Go(func() error {
		out.P95LatencyMS = s.approxP95Latency(egCtx, q)
		return nil
	})

	if err := eg.Wait(); err != nil {
		return out, err
	}
	return out, nil
}

func (s *dbStore) baseQuery(ctx context.Context, q OverviewQuery) *gorm.DB {
	tx := s.db.WithContext(ctx).Model(&Event{}).
		Where("created_at >= ? AND created_at < ?", q.From.UTC(), q.To.UTC())
	if q.UserID != nil {
		tx = tx.Where("user_id = ?", *q.UserID)
	}
	if q.GroupID != nil {
		tx = tx.Where("group_id = ?", *q.GroupID)
	}
	return tx
}

// approxP95Latency approximates p95 latency using the 95th percentile row.
func (s *dbStore) approxP95Latency(ctx context.Context, q OverviewQuery) float64 {
	var count int64
	if err := s.baseQuery(ctx, q).Count(&count).Error; err != nil || count == 0 {
		return 0
	}
	offset := int(float64(count) * 0.95)
	if offset < 0 {
		offset = 0
	}
	var v float64
	s.baseQuery(ctx, q).
		Select("latency_ms").
		Order("latency_ms ASC").
		Offset(offset).
		Limit(1).
		Scan(&v)
	return v
}

func (s *dbStore) StatsTimeseries(ctx context.Context, q TimeseriesQuery) (Timeseries, error) {
	if q.From.IsZero() || q.To.IsZero() {
		return Timeseries{}, ErrTimeRangeRequired
	}
	if q.Granularity != "hour" && q.Granularity != "day" {
		return Timeseries{}, errors.New("invalid granularity: must be 'hour' or 'day'")
	}
	bucketExpr := s.bucketExpr(q.Granularity)

	sql := fmt.Sprintf(`
		SELECT %s AS bucket,
		       COUNT(*) AS request_count,
		       SUM(CASE WHEN status_code>=400 OR error_code!=0 THEN 1 ELSE 0 END) AS error_count,
		       COALESCE(SUM(total_quota), 0) AS total_quota,
		       COALESCE(SUM(input_tokens), 0) AS input_tokens,
		       COALESCE(SUM(output_tokens), 0) AS output_tokens,
		       COALESCE(AVG(latency_ms), 0) AS avg_latency_ms
		FROM request_logs
		WHERE created_at >= ? AND created_at < ?`, bucketExpr)

	args := []any{q.From.UTC(), q.To.UTC()}
	if q.UserID != nil {
		sql += " AND user_id = ?"
		args = append(args, *q.UserID)
	}
	if q.ChannelID != nil {
		sql += " AND channel_id = ?"
		args = append(args, *q.ChannelID)
	}
	if q.Model != "" {
		if strings.ContainsRune(q.Model, '*') {
			sql += " AND client_model LIKE ?"
			args = append(args, strings.ReplaceAll(q.Model, "*", "%"))
		} else {
			sql += " AND client_model = ?"
			args = append(args, q.Model)
		}
	}
	sql += " GROUP BY bucket ORDER BY bucket ASC"

	var rows []TimeseriesPoint
	if err := s.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return Timeseries{}, err
	}
	return Timeseries{Granularity: q.Granularity, Points: rows}, nil
}

func (s *dbStore) bucketExpr(gran string) string {
	if s.driver == "postgres" {
		return fmt.Sprintf("DATE_TRUNC('%s', created_at)", gran)
	}
	if gran == "hour" {
		return "DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')"
	}
	return "DATE_FORMAT(created_at, '%Y-%m-%d')"
}

func (s *dbStore) StatsByModel(ctx context.Context, q GroupQuery) (GroupResult, error) {
	return s.statsBy(ctx, q, "client_model")
}

func (s *dbStore) StatsByChannel(ctx context.Context, q GroupQuery) (GroupResult, error) {
	return s.statsBy(ctx, q, "channel_id")
}

func (s *dbStore) StatsByUser(ctx context.Context, q GroupQuery) (GroupResult, error) {
	return s.statsBy(ctx, q, "user_id")
}

func (s *dbStore) statsBy(ctx context.Context, q GroupQuery, keyCol string) (GroupResult, error) {
	if q.From.IsZero() || q.To.IsZero() {
		return GroupResult{}, ErrTimeRangeRequired
	}
	if q.Limit <= 0 {
		q.Limit = 100
	}
	if q.Limit > 500 {
		q.Limit = 500
	}

	orderBy := "request_count DESC"
	switch q.OrderBy {
	case "quota":
		orderBy = "total_quota DESC"
	case "tokens":
		orderBy = "input_tokens + output_tokens DESC"
	}

	castExpr := s.castToTextExpr(keyCol)
	sql := fmt.Sprintf(`
		SELECT %s AS key,
		       COUNT(*) AS request_count,
		       SUM(CASE WHEN status_code>=400 OR error_code!=0 THEN 1 ELSE 0 END) AS error_count,
		       COALESCE(SUM(total_quota), 0) AS total_quota,
		       COALESCE(SUM(input_tokens), 0) AS input_tokens,
		       COALESCE(SUM(output_tokens), 0) AS output_tokens,
		       COALESCE(AVG(latency_ms), 0) AS avg_latency_ms
		FROM request_logs
		WHERE created_at >= ? AND created_at < ?`,
		castExpr)
	args := []any{q.From.UTC(), q.To.UTC()}
	if q.UserID != nil {
		sql += " AND user_id = ?"
		args = append(args, *q.UserID)
	}
	sql += fmt.Sprintf(" GROUP BY %s ORDER BY %s LIMIT ?", keyCol, orderBy)
	args = append(args, q.Limit)

	var rows []GroupRow
	if err := s.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return GroupResult{}, err
	}

	var total int64
	s.db.WithContext(ctx).Raw(
		fmt.Sprintf(`SELECT COUNT(DISTINCT %s) FROM request_logs WHERE created_at >= ? AND created_at < ?`, keyCol),
		q.From.UTC(), q.To.UTC()).Scan(&total)

	return GroupResult{Rows: rows, Total: total}, nil
}
