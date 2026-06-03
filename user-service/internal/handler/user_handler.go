package handler

import (
	"github.com/DoMinhHHung/user-service/internal/domain/entity"
	"github.com/DoMinhHHung/user-service/internal/dto"
	"github.com/DoMinhHHung/user-service/internal/middleware"
	"github.com/DoMinhHHung/user-service/internal/service"
	"github.com/DoMinhHHung/user-service/pkg/apperr"
	"github.com/DoMinhHHung/user-service/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	svc      *service.UserService
	validate *validator.Validate
}

func NewUser(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc, validate: validator.New()}
}

// GetMyProfile godoc
// @Summary      Lấy profile của tôi
// @Description  Trả về thông tin profile đầy đủ của user đang đăng nhập
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response{data=dto.UserResponse}
// @Failure      401 {object} response.Response
// @Router       /users/me [get]
func (h *UserHandler) GetMyProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.svc.GetMyProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, resp)
}

// UpdateMyProfile godoc
// @Summary      Cập nhật profile
// @Description  Cập nhật thông tin profile. Sau khi điền đủ sẽ set profile_completed=true
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.UpdateProfileRequest true "Profile data"
// @Success      200 {object} response.Response{data=dto.UserResponse}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      409 {object} response.Response "Phone number already in use"
// @Router       /users/me [put]
func (h *UserHandler) UpdateMyProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ValidationError(c, formatValidationError(err))
		return
	}

	userID := middleware.GetUserID(c)
	resp, err := h.svc.UpdateMyProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, resp)
}

// GetUserByID godoc
// @Summary      Lấy thông tin public của user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID"
// @Success      200 {object} response.Response{data=dto.UserResponse}
// @Failure      404 {object} response.Response
// @Router       /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, apperr.ErrInvalidInput)
		return
	}
	resp, err := h.svc.GetUserByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, resp)
}

// UploadAvatar godoc
// @Summary      Upload ảnh đại diện
// @Tags         users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        avatar formData file true "Avatar image (max 5MB, jpg/png/webp)"
// @Success      200 {object} response.Response{data=dto.UserResponse}
// @Failure      400 {object} response.Response
// @Failure      413 {object} response.Response "File too large"
// @Router       /users/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		response.ValidationError(c, "avatar file is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, apperr.ErrInternal)
		return
	}
	defer file.Close()

	userID := middleware.GetUserID(c)
	resp, err := h.svc.UploadAvatar(c.Request.Context(), userID, file, fileHeader)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, resp)
}

// DeleteAvatar godoc
// @Summary      Xóa ảnh đại diện
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response "No avatar to delete"
// @Router       /users/avatar [delete]
func (h *UserHandler) DeleteAvatar(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.svc.DeleteAvatar(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKMessage(c, "Avatar deleted successfully")
}

func formatValidationError(err error) string {
	if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
		return ve[0].Field() + ": " + ve[0].Tag()
	}
	return err.Error()
}

func (h *UserHandler) CreateProfileInternal(c *gin.Context) {
	var req dto.InternalCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ValidationError(c, formatValidationError(err))
		return
	}

	role := entity.UserRole(req.Role)
	if err := h.svc.CreateUserFromEvent(
		c.Request.Context(), req.UserID, req.Email, role,
	); err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, gin.H{"user_id": req.UserID})
}
