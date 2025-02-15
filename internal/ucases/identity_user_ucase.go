package ucases

import (
	"context"
	"net/http"
	"time"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
	"github.com/genefriendway/human-network-iam/packages/utils"
	"github.com/google/uuid"
)

type userUseCase struct {
	userRepo interfaces.IdentityUserRepository
}

func NewIdentityUserUseCase(
	userRepo interfaces.IdentityUserRepository,
) interfaces.IdentityUserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
}

func (u *userUseCase) ChallengeWithPhone(
	ctx context.Context,
	phone string,
) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse) {
	if utils.IsPhoneNumber(phone) {
		return nil, &dto.ErrorDTOResponse{
			Code:    "INVALID_PHONE_NUMBER",
			Message: "Invalid phone number",
			Details: []interface{}{
				map[string]string{
					"field": "phone",
					"error": "Invalid phone number",
				},
			},
		}
	}
	return nil, nil
}

func (u *userUseCase) ChallengeWithEmail(
	ctx context.Context,
	email string,
) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse) {
	if !utils.IsEmail(email) {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_EMAIL",
			Message: "Invalid email",
			Details: []interface{}{
				map[string]string{
					"field": "email",
					"error": "Invalid email",
				},
			},
		}
	}

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Internal server error",
			Details: []interface{}{err},
		}
	}

	if user == nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
			Details: []interface{}{
				map[string]string{
					"field": "user",
					"error": "User not found",
				},
			},
		}
	}

	// Send email with OTP

	// Create challenge session
	session := uuid.New().String()
	// Save challenge session to cache

	// Return challenge session
	return &dto.IdentityUserChallengeDTO{
		SessionID:   session,
		Receiver:    email,
		ChallengeAt: time.Now().Unix(),
	}, nil
}

func (u *userUseCase) ChallengeVerify(
	ctx context.Context,
	sessionID string,
	code string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithGoogle(
	ctx context.Context,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithFacebook(
	ctx context.Context,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithApple(
	ctx context.Context,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) Register(
	ctx context.Context,
	payload dto.IdentityUserRegisterDTO,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogIn(
	ctx context.Context,
	username string,
	password string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogOut(
	ctx context.Context,
) *dto.ErrorDTOResponse {
	return nil
}
