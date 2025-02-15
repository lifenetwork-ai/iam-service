package handlers

import (
	"net/http"

	"github.com/genefriendway/human-network-iam/internal/interfaces"
	httpresponse "github.com/genefriendway/human-network-iam/packages/http/response"
	"github.com/gin-gonic/gin"
)

type userHandler struct {
	ucase interfaces.IdentityUserUseCase
}

func NewIdentityUserHandler(ucase interfaces.IdentityUserUseCase) *userHandler {
	return &userHandler{
		ucase: ucase,
	}
}

// ChallengeWithPhone to login with phone and otp.
// @Summary Login with phone and otp
// @Description Login with phone and otp
// @Tags users
// @Accept json
// @Produce json
// @Param challenge body dto.IdentityChallengeWithPhoneDTO true "challenge payload"
// @Success 200 {object} dto.IdentityUserChallengeDTO "Successful make a challenge with Phone and OTP"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/identity/challenge-with-phone [post]
func (h *userHandler) ChallengeWithPhone(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// ChallengeWithEmail to login with email and otp.
// @Summary Login with email and otp
// @Description Login with email and otp
// @Tags users
// @Accept json
// @Produce json
// @Param challenge body dto.IdentityChallengeWithEmailDTO true "challenge payload"
// @Success 200 {object} dto.IdentityUserChallengeDTO "Successful make a challenge with Email and OTP"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/identity/challenge-with-email [post]
func (h *userHandler) ChallengeWithEmail(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Verify the challenge
// @Summary Verify the challenge
// @Description Verify the challenge
// @Tags users
// @Accept json
// @Produce json
// @Param session_id path string true "session_id"
// @Param code path string true "code"
// @Success 200 {object} dto.IdentityUserAuthDTO "Successful verify the challenge"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/identity/challenge-verify [post]
func (h *userHandler) ChallengeVerify(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)

}

// Login to authenticate user.
// @Summary Authenticate user
// @Description Authenticate user
// @Tags users
// @Accept json
// @Produce json
// @Param login body dto.IdentityUserLoginDTO true "login payload"
// @Success 200 {object} dto.IdentityUserAuthDTO "Successful authenticate user"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/identity/login [post]
func (h *userHandler) Login(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Login with Google to authenticate user.
// @Summary Authenticate user with Google
// @Description Authenticate user with Google
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} dto.IdentityUserAuthDTO "Successful authenticate user with Google"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/identity/login-with-google [post]
func (h *userHandler) LoginWithGoogle(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Login with Facebook to authenticate user.
// @Summary Authenticate user with Facebook
// @Description Authenticate user with Facebook
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} dto.IdentityUserAuthDTO "Successful authenticate user with Facebook"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/identity/login-with-facebook [post]
func (h *userHandler) LoginWithFacebook(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Login with Apple to authenticate user.
// @Summary Authenticate user with Apple
// @Description Authenticate user with Apple
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} dto.IdentityUserAuthDTO "Successful authenticate user with Apple"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/identity/login-with-apple [post]
func (h *userHandler) LoginWithApple(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Refresh to refresh token.
// @Summary Refresh token
// @Description Refresh token
// @Tags users
// @Accept json
// @Produce json
// @Param refresh_token body dto.IdentityRefreshTokenDTO true "refresh token payload"
// @Success 200 {object} dto.IdentityUserAuthDTO "Successful refresh token"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/identity/refresh-token [post]
func (h *userHandler) RefreshToken(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Me to get user profile.
// @Summary Get user profile
// @Description Get user profile
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} dto.IdentityUserDTO "Successful get user profile"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/identity/me [get]
func (h *userHandler) Me(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// Logout to de-authenticate user.
// @Summary De-authenticate user
// @Description De-authenticate user
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse "Successful de-authenticate user"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/identity/logout [post]
func (h *userHandler) Logout(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}
