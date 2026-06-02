// internal/adapter/handler/session_handler.go
package handler

import (
	"github.com/DoMinhHHung/auth-service/internal/application/usecase"
	"github.com/DoMinhHHung/auth-service/pkg/response"
	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	sessionUC *usecase.SessionUseCase
}

func NewSessionHandler(sessionUC *usecase.SessionUseCase) *SessionHandler {
	return &SessionHandler{sessionUC: sessionUC}
}

// Logout godoc
// @Summary      Logout current session
// @Tags         sessions
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Router       /auth/logout [post]
func (h *SessionHandler) Logout(c *gin.Context) {
	sessionID := getSessionID(c)
	if err := h.sessionUC.Logout(c.Request.Context(), sessionID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKMessage(c, "Logged out successfully.")
}

// LogoutAll godoc
// @Summary      Logout from all devices
// @Tags         sessions
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Router       /auth/logout/all [post]
func (h *SessionHandler) LogoutAll(c *gin.Context) {
	userID := getUserID(c)
	if err := h.sessionUC.LogoutAll(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKMessage(c, "Logged out from all devices.")
}

// GetActiveSessions godoc
// @Summary      Get all active sessions
// @Tags         sessions
// @Security     BearerAuth
// @Success      200 {object} response.Response{data=[]dto.SessionResponse}
// @Router       /auth/sessions [get]
func (h *SessionHandler) GetActiveSessions(c *gin.Context) {
	userID := getUserID(c)
	sessions, err := h.sessionUC.GetActiveSessions(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, sessions)
}

// RevokeSession godoc
// @Summary      Revoke specific session
// @Tags         sessions
// @Security     BearerAuth
// @Param        id path string true "Session ID"
// @Success      200 {object} response.Response
// @Router       /auth/sessions/{id}/revoke [delete]
func (h *SessionHandler) RevokeSession(c *gin.Context) {
	userID := getUserID(c)
	sessionID := c.Param("id")

	if err := h.sessionUC.RevokeSession(c.Request.Context(), userID, sessionID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKMessage(c, "Session revoked.")
}

func getUserID(c *gin.Context) string {
	id, _ := c.Get("user_id")
	s, _ := id.(string)
	return s
}

func getSessionID(c *gin.Context) string {
	id, _ := c.Get("session_id")
	s, _ := id.(string)
	return s
}
