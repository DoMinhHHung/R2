package handler

import (
	"github.com/DoMinhHHung/auth-service/internal/application/dto"
	"github.com/DoMinhHHung/auth-service/internal/application/usecase"
	"github.com/DoMinhHHung/auth-service/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AdminHandler struct {
	adminLoginUC *usecase.AdminLoginUseCase
	adminUC      *usecase.AdminUseCase
	validate     *validator.Validate
}

func NewAdminHandler(
	adminLoginUC *usecase.AdminLoginUseCase,
	adminUC *usecase.AdminUseCase,
) *AdminHandler {
	return &AdminHandler{
		adminLoginUC: adminLoginUC,
		adminUC:      adminUC,
		validate:     validator.New(),
	}
}

// Login godoc
// @Summary      Admin login
// @Description  Đăng nhập dành riêng cho admin. Chỉ account role=ADMIN mới thành công.
// @Description  Rate limit: 3 requests / 5 minutes per IP.
// @Description  Dùng cùng error code cho "sai password" và "không phải admin" để tránh role enumeration.
// @Tags         admin-auth
// @Accept       json
// @Produce      json
// @Param        request body dto.AdminLoginRequest true "Admin credentials"
// @Success      200 {object} response.Response{data=dto.TokenResponse}
// @Failure      400 {object} response.Response "Validation error"
// @Failure      401 {object} response.Response "Invalid credentials"
// @Failure      429 {object} response.Response "Rate limit exceeded"
// @Router       /auth/admin/login [post]
func (h *AdminHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ValidationError(c, formatValidationError(err))
		return
	}

	lctx := usecase.LoginContext{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	tokens, err := h.adminLoginUC.Login(c.Request.Context(), &req, lctx)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, tokens)
}

// CreateAdmin godoc
// @Summary      Tạo tài khoản admin mới
// @Description  Chỉ admin đang active mới được tạo admin khác. Không cần OTP.
// @Description  Password yêu cầu tối thiểu 12 ký tự.
// @Tags         admin-auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateAdminRequest true "New admin info"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response "Validation error"
// @Failure      401 {object} response.Response "Unauthorized"
// @Failure      403 {object} response.Response "Not an admin"
// @Failure      409 {object} response.Response "Email already exists"
// @Router       /auth/admin/users [post]
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req dto.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ValidationError(c, formatValidationError(err))
		return
	}

	createdByID := getUserID(c) // lấy từ JWT context set bởi RequireAuth

	if err := h.adminUC.CreateAdmin(c.Request.Context(), &req, createdByID); err != nil {
		response.Error(c, err)
		return
	}

	// Trả về 201 Created, không trả về password hay sensitive data
	response.Created(c, gin.H{"message": "Admin account created successfully"})
}
