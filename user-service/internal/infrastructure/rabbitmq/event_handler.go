package rabbitmq

import (
	"context"

	"github.com/DoMinhHHung/user-service/internal/domain/entity"
	"github.com/DoMinhHHung/user-service/internal/service"
)

type userEventHandler struct {
	userSvc *service.UserService
}

func NewUserEventHandler(svc *service.UserService) EventHandler {
	return &userEventHandler{userSvc: svc}
}

func (h *userEventHandler) HandleUserCreated(ctx context.Context, event UserCreatedEvent) error {
	role := entity.RoleTenant
	if event.Role != "" {
		role = entity.UserRole(event.Role)
	}
	return h.userSvc.CreateUserFromEvent(ctx, event.UserID, event.Email, role)
}
