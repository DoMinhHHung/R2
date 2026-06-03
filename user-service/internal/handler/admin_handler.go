package handler

import (
	"strconv"

	"github.com/DoMinhHHung/user-service/internal/domain/entity"
	"github.com/DoMinhHHung/user-service/internal/domain/port"
	"github.com/DoMinhHHung/user-service/internal/middleware"
	"github.com/DoMinhHHung/user-service/internal/service"
	"github.com/DoMinhHHung/user-service/pkg/apperr"
	"github.com/DoMinhHHung/user-service/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	svc *service.UserService
}

func NewAdmin(svc *service.UserService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// ListUsers godoc
// @Summary      Danh sách tất cả users
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        page    query int    false "Page number (default: 1)"
// @Param        limit   query int    false "Page size (default: 20, max: 100)"
// @Param        status  query string false "Filter by status: ACTIVE, BANNED"
// @Param        search  query string false "Search by email or name"
// @Success      200 {object} response.Response{data=dto.PaginatedResponse}
// @Failure      403 {object} response.Response
// @Router       /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := port.UserFilter{Page: page, Limit: limit}

	if s := c.Query("status"); s != "" {
		status := entity.UserStatus(s)
		filter.Status = &status
	}
	if s := c.Query("search"); s != "" {
		filter.Search = s
	}

	result, err := h.svc.AdminListUsers(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// GetUser godoc
// @Summary      Lấy thông tin đầy đủ của user (admin)
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID"
// @Success      200 {object} response.Response{data=dto.AdminUserResponse}
// @Failure      404 {object} response.Response
// @Router       /admin/users/{id} [get]
func (h *AdminHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.svc.AdminGetUser(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, resp)
}

// BanUser godoc
// @Summary      Ban user
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response "Already banned or self-ban"
// @Failure      404 {object} response.Response
// @Router       /admin/users/{id}/ban [patch]
func (h *AdminHandler) BanUser(c *gin.Context) {
	targetID := c.Param("id")
	adminID := middleware.GetUserID(c)
	if targetID == "" {
		response.Error(c, apperr.ErrInvalidInput)
		return
	}
	if err := h.svc.BanUser(c.Request.Context(), targetID, adminID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKMessage(c, "User banned successfully")
}

// UnbanUser godoc
// @Summary      Unban user
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID"
// @Success      200 {object} response.Response
// @Router       /admin/users/{id}/unban [patch]
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	targetID := c.Param("id")
	if err := h.svc.UnbanUser(c.Request.Context(), targetID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKMessage(c, "User unbanned successfully")
}
