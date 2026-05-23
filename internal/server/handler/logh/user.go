package logh

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ilog "github.com/ijry/pro-api/internal/log"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// User handles user-facing log endpoints.
type User struct {
	log ilog.Store
}

// NewUser constructs a User handler.
func NewUser(s ilog.Store) *User {
	return &User{log: s}
}

// Register attaches user routes to the router group.
func (h *User) Register(g *gin.RouterGroup) {
	g.GET("/logs/requests", h.QueryRequests)
	g.GET("/logs/errors", h.QueryErrors)
}

func (h *User) QueryRequests(c *gin.Context) {
	userID := getUserID(c)
	q, err := parseQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	// Force inject user_id (override any client-supplied value)
	q.UserID = &userID
	res, err := h.log.Query(c, q)
	if err != nil {
		middleware.SetErr(c, mapStoreErr(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *User) QueryErrors(c *gin.Context) {
	userID := getUserID(c)
	q, err := parseErrorQuery(c)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	// Force inject user_id
	q.UserID = &userID
	res, err := h.log.QueryErrors(c, q)
	if err != nil {
		middleware.SetErr(c, mapStoreErr(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

// getUserID extracts the user ID from gin context (set by SessionAuth middleware).
func getUserID(c *gin.Context) int64 {
	if v, ok := c.Get(middleware.CtxKeyUserID); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}
