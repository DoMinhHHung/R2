package handler

import (
	"github.com/DoMinhHHung/auth-service/internal/application/dto"
	"github.com/DoMinhHHung/auth-service/internal/application/usecase"
	"github.com/DoMinhHHung/auth-service/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	signupUC   *usecase.SignupUseCase
	loginUC    *usecase.LoginUseCase
	tokenUC    *usecase.TokenUseCase
	passwordUC *usecase.PasswordUseCase
	validate   *validator.Validate
}

func NewAuthHandler(
	signupUC *usecase.SignupUseCase,
	loginUC *usecase.LoginUseCase,
	tokenUC *usecase.TokenUseCase,
	passwordUC *usecase.PasswordUseCase,
) *AuthHandler {
	return &AuthHandler{
		signupUC:   signupUC,
		loginUC:    loginUC,
		tokenUC:    tokenUC,
		passwordUC: passwordUC,
		validate:   validator.New(),
	}
}

// Signup godoc
// @Summary      Initiate signup
// @Description  Start signup process — sends OTP to email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.SignupRequest true "Signup request"
// @Success      200 {object} response.Response{message=string}
// @Failure      400 {object} response.Response
// @Failure      409 {object} response.Response "Email already exists"
// @Failure      429 {object} response.Response "Rate limit exceeded"
// @Router       /auth/signup [post]
func (h *AuthHandler) Signup(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ValidationError(c, formatValidationError(err))
		return
	}

	if err := h.signupUC.InitiateSignup(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.OKMessage(c, "OTP sent to your email. Valid for 5 minutes.")
}

// VerifySignupOTP godoc
// @Summary      Verify signup OTP
// @Description  Verify OTP and create account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.VerifyOTPRequest true "Verify OTP request"
// @Success      200 {object} response.Response{message=string}
// @Failure      400 {object} response.Response
// @Failure      410 {object} response.Response "OTP expired"
// @Router       /auth/signup/verify [post]
func (h *AuthHandler) VerifySignupOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ValidationError(c, formatValidationError(err))
		return
	}

	if err := h.signupUC.VerifyOTP(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.OKMessage(c, "Account created successfully. Please login.")
}

// ResendSignupOTP godoc
// @Summary      Resend signup OTP
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResendOTPRequest true "Email"
// @Success      200 {object} response.Response
// @Router       /auth/signup/resend [post]
func (h *AuthHandler) ResendSignupOTP(c *gin.Context) {
	var req dto.ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := h.signupUC.ResendOTP(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}
	response.OKMessage(c, "New OTP sent.")
}

// Login godoc
// @Summary      Login
// @Description  Login for TENANT, LANDLORD, or ADMIN
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login request"
// @Success      200 {object} response.Response{data=dto.TokenResponse}
// @Failure      401 {object} response.Response
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
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

	tokens, err := h.loginUC.Login(c.Request.Context(), &req, lctx)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, tokens)
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshTokenRequest true "Refresh token"
// @Success      200 {object} response.Response{data=dto.TokenResponse}
// @Failure      401 {object} response.Response
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	tokens, err := h.tokenUC.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, tokens)
}

// ForgotPassword godoc
// @Summary      Forgot password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "Email"
// @Success      200 {object} response.Response
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Luôn return 200 dù email có tồn tại hay không (Email Enumeration Protection)
	_ = h.passwordUC.ForgotPassword(c.Request.Context(), &req)
	response.OKMessage(c, "If your email is registered, you will receive an OTP shortly.")
}

// ResendForgotOTP godoc
// @Summary      Resend forgot password OTP
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "Email"
// @Success      200 {object} response.Response
// @Router       /auth/forgot-password/resend [post]
func (h *AuthHandler) ResendForgotOTP(c *gin.Context) {
	h.ForgotPassword(c)
}

// VerifyRecoveryOTP godoc
// @Summary      Verify recovery OTP
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.VerifyRecoveryOTPRequest true "Recovery OTP"
// @Success      200 {object} response.Response{data=map[string]string}
// @Router       /auth/forgot-password/verify [post]
func (h *AuthHandler) VerifyRecoveryOTP(c *gin.Context) {
	var req dto.VerifyRecoveryOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	resetToken, err := h.passwordUC.VerifyRecoveryOTP(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, gin.H{"reset_token": resetToken})
}

// ResetPassword godoc
// @Summary      Reset password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Reset password"
// @Success      200 {object} response.Response
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.passwordUC.ResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}
	response.OKMessage(c, "Password reset successfully.")
}

func formatValidationError(err error) string {
	if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
		return ve[0].Field() + ": " + ve[0].Tag()
	}
	return err.Error()
}
