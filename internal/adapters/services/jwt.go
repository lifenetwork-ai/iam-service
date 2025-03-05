package services

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	OrganizationId string `json:"organization_id"`
	UserId         string `json:"user_id"`
	UserName       string `json:"user_name"`
	jwt.RegisteredClaims
}

type JWTToken struct {
	AccessToken        string    `json:"access_token"`
	RefreshToken       string    `json:"refresh_token"`
	AccessTokenExpiry  int64     `json:"access_token_expiry"`
	RefreshTokenExpiry int64     `json:"refresh_token_expiry"`
	ClaimAt            time.Time `json:"claim_at"`
}

type JWTService interface {
	GenerateToken(ctx context.Context, claims JWTClaims) (*JWTToken, error)
	ValidateToken(ctx context.Context, token string) (*JWTClaims, error)
}

type JWTServiceSetting struct {
	Secret          string
	AccessLifetime  int64
	RefreshLifetime int64
}

func NewJWTService(
	secret string,
	accessLifetime int64,
	refreshLifetime int64,
) JWTService {
	return &JWTServiceSetting{
		Secret:          secret,
		AccessLifetime:  accessLifetime,
		RefreshLifetime: refreshLifetime,
	}
}

func (c *JWTServiceSetting) GenerateToken(
	ctx context.Context,
	claims JWTClaims,
) (*JWTToken, error) {
	accessClaims := &JWTClaims{
		OrganizationId: claims.OrganizationId,
		UserId:         claims.UserId,
		UserName:       claims.UserName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(c.AccessLifetime) * time.Second)),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(c.Secret))
	if err != nil {
		return nil, err
	}

	refreshClaims := &JWTClaims{
		OrganizationId: claims.OrganizationId,
		UserId:         claims.UserId,
		UserName:       claims.UserName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(c.RefreshLifetime) * time.Second)),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(c.Secret))
	if err != nil {
		return nil, err
	}

	jwtToken := &JWTToken{
		AccessToken:        signedAccessToken,
		RefreshToken:       signedRefreshToken,
		AccessTokenExpiry:  accessClaims.ExpiresAt.Unix(),
		RefreshTokenExpiry: refreshClaims.ExpiresAt.Unix(),
		ClaimAt:            time.Now(),
	}

	return jwtToken, nil
}

func (c *JWTServiceSetting) ValidateToken(
	ctx context.Context,
	token string,
) (*JWTClaims, error) {
	claims := &JWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.Secret), nil
	})

	if err != nil || !parsedToken.Valid {
		return nil, err
	}

	return claims, nil
}
