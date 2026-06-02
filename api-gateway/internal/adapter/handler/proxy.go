package handler

import (
	"github.com/DoMinhHHung/R2/internal/usecase/proxy"
	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	uc *proxy.UseCase
}

func NewProxy(uc *proxy.UseCase) *ProxyHandler {
	return &ProxyHandler{uc: uc}
}

func (h *ProxyHandler) Handle(c *gin.Context) {
	route, ok := h.uc.Resolve(c.Request.URL.Path)
	if !ok {
		c.JSON(404, gin.H{"error": "route_not_found", "path": c.Request.URL.Path})
		return
	}

	if err := h.uc.Forward(c.Request.Context(), route, c.Writer, c.Request); err != nil {
		if c.Writer.Written() {
			return
		}
		c.JSON(502, gin.H{"error": "upstream_error", "detail": err.Error()})
	}
}
