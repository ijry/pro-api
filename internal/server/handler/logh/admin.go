// Package logh provides HTTP handlers for log querying and statistics.
package logh

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	ilog "github.com/ijry/pro-api/internal/log"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// ChannelNameLooker is the minimal interface to look up channel names.
type ChannelNameLooker interface {
	NamesByIDs(ctx context.Context, ids []int64) map[int64]string
}

// UserNameLooker is the minimal interface to look up usernames.
type UserNameLooker interface {
	UsernamesByIDs(ctx context.Context, ids []int64) map[int64]string
}

// Admin handles admin log and stats endpoints.
type Admin struct {
	log        ilog.Store
	channelSvc ChannelNameLooker
	userSvc    UserNameLooker
}

// NewAdmin constructs an Admin handler.
func NewAdmin(s ilog.Store, ch ChannelNameLooker, u UserNameLooker) *Admin {
	return &Admin{log: s, channelSvc: ch, userSvc: u}
}

// Register attaches admin routes to the router group.
func (h *Admin) Register(g *gin.RouterGroup) {
	g.GET("/logs/requests", h.QueryRequests)
	g.GET("/logs/errors", h.QueryErrors)
	g.GET("/logs/audit", h.QueryAudits)
	g.GET("/stats/overview", h.StatsOverview)
	g.GET("/stats/timeseries", h.StatsTimeseries)
	g.GET("/stats/by_model", h.StatsByModel)
	g.GET("/stats/by_channel", h.StatsByChannel)
	g.GET("/stats/by_user", h.StatsByUser)
	g.GET("/stats/export", h.ExportCSV)
}

func (h *Admin) QueryRequests(c *gin.Context) {
	q, err := parseQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	res, err := h.log.Query(c, q)
	if err != nil {
		middleware.SetErr(c, mapStoreErr(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Admin) QueryErrors(c *gin.Context) {
	q, err := parseErrorQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	res, err := h.log.QueryErrors(c, q)
	if err != nil {
		middleware.SetErr(c, mapStoreErr(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Admin) QueryAudits(c *gin.Context) {
	q, err := parseAuditQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	res, err := h.log.QueryAudits(c, q)
	if err != nil {
		middleware.SetErr(c, mapStoreErr(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Admin) StatsOverview(c *gin.Context) {
	q, err := parseOverviewQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	res, err := h.log.StatsOverview(c, q)
	if err != nil {
		middleware.SetErr(c, mapStoreErr(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Admin) StatsTimeseries(c *gin.Context) {
	q, err := parseTimeseriesQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	res, err := h.log.StatsTimeseries(c, q)
	if err != nil {
		middleware.SetErr(c, mapStoreErr(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Admin) StatsByModel(c *gin.Context) {
	q, err := parseGroupQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	res, err := h.log.StatsByModel(c, q)
	if err != nil {
		middleware.SetErr(c, mapStoreErr(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Admin) StatsByChannel(c *gin.Context) {
	q, err := parseGroupQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	res, err := h.log.StatsByChannel(c, q)
	if err != nil {
		middleware.SetErr(c, mapStoreErr(err))
		return
	}
	// Enrich labels
	if h.channelSvc != nil {
		ids := make([]int64, 0, len(res.Rows))
		for _, r := range res.Rows {
			if id, e := strconv.ParseInt(r.Key, 10, 64); e == nil {
				ids = append(ids, id)
			}
		}
		nameByID := h.channelSvc.NamesByIDs(c, ids)
		for i := range res.Rows {
			if id, e := strconv.ParseInt(res.Rows[i].Key, 10, 64); e == nil {
				if name, ok := nameByID[id]; ok {
					res.Rows[i].Label = name
				}
			}
		}
	}
	c.JSON(http.StatusOK, res)
}

func (h *Admin) StatsByUser(c *gin.Context) {
	q, err := parseGroupQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	res, err := h.log.StatsByUser(c, q)
	if err != nil {
		middleware.SetErr(c, mapStoreErr(err))
		return
	}
	// Enrich labels
	if h.userSvc != nil {
		ids := make([]int64, 0, len(res.Rows))
		for _, r := range res.Rows {
			if id, e := strconv.ParseInt(r.Key, 10, 64); e == nil {
				ids = append(ids, id)
			}
		}
		nameByID := h.userSvc.UsernamesByIDs(c, ids)
		for i := range res.Rows {
			if id, e := strconv.ParseInt(res.Rows[i].Key, 10, 64); e == nil {
				if name, ok := nameByID[id]; ok {
					res.Rows[i].Label = name
				}
			}
		}
	}
	c.JSON(http.StatusOK, res)
}

func (h *Admin) ExportCSV(c *gin.Context) {
	q, err := parseQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	fn := fmt.Sprintf("request_logs_%s_%s.csv",
		q.From.Format("20060102"), q.To.Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fn))
	c.Header("X-Accel-Buffering", "no")
	n, err := h.log.ExportCSV(c, q, c.Writer)
	if err != nil {
		return
	}
	c.Header("X-Total-Rows", strconv.FormatInt(n, 10))
}

// --- Query parsers ---

func parseQuery(c *gin.Context) (ilog.Query, error) {
	var q ilog.Query
	from, to, err := parseTimeRange(c)
	if err != nil {
		return q, err
	}
	q.From, q.To = from, to

	if s := c.Query("user_id"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return q, fmt.Errorf("user_id: %w", err)
		}
		q.UserID = &v
	}
	if s := c.Query("token_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.TokenID = &v
	}
	if s := c.Query("channel_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.ChannelID = &v
	}
	if s := c.Query("event_type"); s != "" {
		v, _ := strconv.Atoi(s)
		et := int8(v)
		q.EventType = &et
	}
	if s := c.Query("status_code"); s != "" {
		v, _ := strconv.Atoi(s)
		q.StatusCode = &v
	}
	q.ClientModel = c.Query("client_model")
	q.TraceID = c.Query("trace_id")

	if s := c.Query("page_size"); s != "" {
		v, _ := strconv.Atoi(s)
		q.PageSize = v
	}
	if s := c.Query("cursor"); s != "" {
		cur, err := ilog.ParseCursor(s)
		if err != nil {
			return q, fmt.Errorf("cursor: %w", err)
		}
		q.Cursor = cur
	}
	return q, nil
}

func parseErrorQuery(c *gin.Context) (ilog.ErrorQuery, error) {
	var q ilog.ErrorQuery
	from, to, err := parseTimeRange(c)
	if err != nil {
		return q, err
	}
	q.From, q.To = from, to

	if s := c.Query("user_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.UserID = &v
	}
	if s := c.Query("token_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.TokenID = &v
	}
	if s := c.Query("channel_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.ChannelID = &v
	}
	if s := c.Query("error_code"); s != "" {
		v, _ := strconv.Atoi(s)
		q.ErrorCode = &v
	}
	q.ErrorType = c.Query("error_type")
	q.TraceID = c.Query("trace_id")

	if s := c.Query("page_size"); s != "" {
		v, _ := strconv.Atoi(s)
		q.PageSize = v
	}
	if s := c.Query("cursor"); s != "" {
		cur, err := ilog.ParseCursor(s)
		if err != nil {
			return q, fmt.Errorf("cursor: %w", err)
		}
		q.Cursor = cur
	}
	return q, nil
}

func parseAuditQuery(c *gin.Context) (ilog.AuditQuery, error) {
	var q ilog.AuditQuery
	from, to, err := parseTimeRange(c)
	if err != nil {
		return q, err
	}
	q.From, q.To = from, to

	if s := c.Query("actor_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.ActorID = &v
	}
	q.Action = c.Query("action")
	q.TargetType = c.Query("target_type")
	if s := c.Query("target_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.TargetID = &v
	}
	if s := c.Query("page_size"); s != "" {
		v, _ := strconv.Atoi(s)
		q.PageSize = v
	}
	if s := c.Query("cursor"); s != "" {
		cur, err := ilog.ParseCursor(s)
		if err != nil {
			return q, fmt.Errorf("cursor: %w", err)
		}
		q.Cursor = cur
	}
	return q, nil
}

func parseOverviewQuery(c *gin.Context) (ilog.OverviewQuery, error) {
	var q ilog.OverviewQuery
	from, to, err := parseTimeRange(c)
	if err != nil {
		return q, err
	}
	q.From, q.To = from, to
	if s := c.Query("user_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.UserID = &v
	}
	if s := c.Query("group_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.GroupID = &v
	}
	return q, nil
}

func parseTimeseriesQuery(c *gin.Context) (ilog.TimeseriesQuery, error) {
	var q ilog.TimeseriesQuery
	from, to, err := parseTimeRange(c)
	if err != nil {
		return q, err
	}
	q.From, q.To = from, to
	q.Granularity = c.Query("granularity")
	if q.Granularity == "" {
		q.Granularity = "hour"
	}
	if s := c.Query("user_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.UserID = &v
	}
	if s := c.Query("channel_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.ChannelID = &v
	}
	q.Model = c.Query("model")
	return q, nil
}

func parseGroupQuery(c *gin.Context) (ilog.GroupQuery, error) {
	var q ilog.GroupQuery
	from, to, err := parseTimeRange(c)
	if err != nil {
		return q, err
	}
	q.From, q.To = from, to
	if s := c.Query("user_id"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 64)
		q.UserID = &v
	}
	if s := c.Query("limit"); s != "" {
		v, _ := strconv.Atoi(s)
		q.Limit = v
	}
	q.OrderBy = c.Query("order_by")
	return q, nil
}

func parseTimeRange(c *gin.Context) (from, to time.Time, err error) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		return from, to, errors.New("from and to are required")
	}
	from, err = time.Parse(time.RFC3339, fromStr)
	if err != nil {
		return from, to, fmt.Errorf("invalid from: %w", err)
	}
	to, err = time.Parse(time.RFC3339, toStr)
	if err != nil {
		return from, to, fmt.Errorf("invalid to: %w", err)
	}
	from = from.UTC()
	to = to.UTC()
	if !to.After(from) {
		return from, to, errors.New("to must be after from")
	}
	return from, to, nil
}

func mapStoreErr(err error) *apierr.Error {
	switch {
	case errors.Is(err, ilog.ErrTimeRangeRequired):
		return apierr.New(apierr.CodeInvalidParam, "time range (from/to) required")
	case errors.Is(err, ilog.ErrTimeRangeTooWide):
		return apierr.New(apierr.CodeInvalidParam, "time range too wide (max 31 days)")
	case errors.Is(err, ilog.ErrInvalidCursor):
		return apierr.New(apierr.CodeInvalidParam, "invalid cursor")
	case errors.Is(err, ilog.ErrCHNotImplemented):
		return apierr.New(apierr.CodeInternal, "clickhouse backend not implemented")
	default:
		return apierr.New(apierr.CodeDatabase, err.Error())
	}
}

// Ensure unused import used.
var _ = strings.TrimSpace
