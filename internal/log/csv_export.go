package log

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const defaultExportMaxRows = 1_000_000

var csvHeaders = []string{
	"id", "created_at", "user_id", "token_id", "channel_id",
	"event_type", "client_model", "upstream_model", "protocol", "endpoint",
	"ip", "status_code", "latency_ms", "ttft_ms", "stream",
	"input_tokens", "output_tokens", "cached_tokens", "reasoning_tokens",
	"total_quota", "billing_input_ratio", "billing_output_ratio", "billing_group_ratio",
	"error_code", "error_msg", "trace_id",
}

func (s *dbStore) ExportCSV(ctx context.Context, q Query, w io.Writer) (int64, error) {
	if q.From.IsZero() || q.To.IsZero() {
		return 0, ErrTimeRangeRequired
	}
	maxRows := int64(defaultExportMaxRows)

	cw := csv.NewWriter(w)
	defer cw.Flush()

	if err := cw.Write(csvHeaders); err != nil {
		return 0, err
	}
	cw.Flush()
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	var written int64
	q.PageSize = 1000
	q.Cursor = nil

	for {
		page, err := s.Query(ctx, q)
		if err != nil {
			return written, err
		}
		for _, e := range page.Items {
			if written >= maxRows {
				_ = cw.Write([]string{"_TRUNCATED", "", "", "", "", "", "", "", "", "",
					"", "", "", "", "", "", "", "", "", "", "", "", "", "", "row_limit_exceeded", ""})
				cw.Flush()
				return written, nil
			}
			if err := cw.Write(eventToCSVRow(e)); err != nil {
				return written, err
			}
			written++
		}
		cw.Flush()
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		if page.NextCursor == nil {
			break
		}
		q.Cursor = page.NextCursor
	}
	return written, nil
}

func eventToCSVRow(e Event) []string {
	return []string{
		strconv.FormatInt(e.ID, 10),
		e.CreatedAt.UTC().Format(time.RFC3339Nano),
		strconv.FormatInt(e.UserID, 10),
		strconv.FormatInt(e.TokenID, 10),
		nullableInt64(e.ChannelID),
		strconv.Itoa(int(e.EventType)),
		e.ClientModel,
		e.UpstreamModel,
		e.Protocol,
		e.Endpoint,
		e.IP,
		strconv.Itoa(e.StatusCode),
		strconv.Itoa(e.LatencyMS),
		strconv.Itoa(e.TTFTMS),
		strconv.FormatBool(e.Stream),
		strconv.Itoa(e.InputTokens),
		strconv.Itoa(e.OutputTokens),
		strconv.Itoa(e.CachedTokens),
		strconv.Itoa(e.ReasoningTokens),
		strconv.FormatInt(e.TotalQuota, 10),
		fmt.Sprintf("%.4f", e.BillingInputRatio),
		fmt.Sprintf("%.4f", e.BillingOutputRatio),
		fmt.Sprintf("%.4f", e.BillingGroupRatio),
		strconv.Itoa(e.ErrorCode),
		e.ErrorMsg,
		e.TraceID,
	}
}

func nullableInt64(p *int64) string {
	if p == nil {
		return ""
	}
	return strconv.FormatInt(*p, 10)
}
