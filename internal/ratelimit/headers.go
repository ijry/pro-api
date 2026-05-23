package ratelimit

import (
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// WriteHeaders 把限流响应头写到 gin.Context。
//
// 被拒(d.Allowed=false):写 X-RateLimit-Limit-{Suffix}, X-RateLimit-Remaining-{Suffix}=0,
// X-RateLimit-Reset, Retry-After(≥1)。
//
// 通过(d.Allowed=true):写最紧张维度的 Limit / Remaining / Reset(不写 Retry-After)。
func WriteHeaders(c *gin.Context, d Decision) {
	suffix := d.Dimension.HeaderSuffix()
	if suffix == "" {
		return
	}
	if d.Allowed {
		c.Header("X-RateLimit-Limit-"+suffix, strconv.Itoa(d.Limit))
		c.Header("X-RateLimit-Remaining-"+suffix, strconv.Itoa(d.Remaining))
		if !d.Reset.IsZero() {
			c.Header("X-RateLimit-Reset", strconv.FormatInt(d.Reset.Unix(), 10))
		}
		return
	}
	// denied
	c.Header("X-RateLimit-Limit-"+suffix, strconv.Itoa(d.Limit))
	c.Header("X-RateLimit-Remaining-"+suffix, "0")
	if !d.Reset.IsZero() {
		c.Header("X-RateLimit-Reset", strconv.FormatInt(d.Reset.Unix(), 10))
	}
	retryAfter := 1
	if !d.Reset.IsZero() {
		s := time.Until(d.Reset).Seconds()
		ra := int(math.Ceil(s))
		if ra < 1 {
			ra = 1
		}
		retryAfter = ra
	}
	c.Header("Retry-After", strconv.Itoa(retryAfter))
}
