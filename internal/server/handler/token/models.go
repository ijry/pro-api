package token

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/server/middleware"
	tokensvc "github.com/ijry/pro-api/internal/token"
	"github.com/ijry/pro-api/pkg/apierr"
)

// ModelObject 是 OpenAI /v1/models 协议里的单个 model 对象。
type ModelObject struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ModelsResponse 是 GET /v1/models 的完整响应体。
type ModelsResponse struct {
	Object string        `json:"object"`
	Data   []ModelObject `json:"data"`
}

// ModelsHandler 提供 GET /v1/models。
//
// 数据来源:
//
//   - ModelLister.ActiveModels(渠道聚合)
//   - 用 token.View.AllowedModels 过滤
//   - ModelLister.ModelInfo 补 created/owned_by(缺失时默认 0/"system")
type ModelsHandler struct {
	lister tokensvc.ModelLister
}

// NewModelsHandler 构造。lister 可以为 nil(渠道模块未装配时返回空列表)。
func NewModelsHandler(lister tokensvc.ModelLister) *ModelsHandler {
	return &ModelsHandler{lister: lister}
}

// Register 把 GET /v1/models 挂到 group(group 已经过 TokenAuth)。
func (h *ModelsHandler) Register(g *gin.RouterGroup) {
	g.GET("/models", h.List)
}

// List 处理 GET /v1/models。
func (h *ModelsHandler) List(c *gin.Context) {
	view, ok := tokensvc.FromContext(c)
	if !ok || view == nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidToken, "missing token in context"))
		return
	}
	if h.lister == nil {
		// 渠道模块未装配时返回空列表(避免 500)
		c.JSON(http.StatusOK, ModelsResponse{Object: "list", Data: []ModelObject{}})
		return
	}
	models := h.lister.ActiveModels(c.Request.Context())

	// 过滤白名单
	filtered := make([]string, 0, len(models))
	for _, m := range models {
		if tokensvc.ModelInAllowList(view, m) {
			filtered = append(filtered, m)
		}
	}
	sort.Strings(filtered)

	// 转响应
	data := make([]ModelObject, 0, len(filtered))
	for _, m := range filtered {
		meta, _ := h.lister.ModelInfo(m)
		if meta.OwnedBy == "" {
			meta.OwnedBy = "system"
		}
		data = append(data, ModelObject{
			ID:      m,
			Object:  "model",
			Created: meta.Created,
			OwnedBy: meta.OwnedBy,
		})
	}
	c.JSON(http.StatusOK, ModelsResponse{Object: "list", Data: data})
}
