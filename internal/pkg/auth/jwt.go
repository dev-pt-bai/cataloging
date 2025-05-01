package auth

import (
	"time"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(user *model.User, tokenExpiry time.Duration, secret string) (*model.Auth, *errors.Error) {
	if user == nil {
		return nil, errors.New(errors.UserNotFound)
	}

	accessExpiredAt := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	refreshExpiredAt := accessExpiredAt
	if tokenExpiry > 0 {
		now := time.Now()
		accessExpiredAt = now.Add(time.Hour * tokenExpiry).Unix()
		refreshExpiredAt = now.Add(10 * time.Hour * tokenExpiry).Unix()
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, (model.Auth{
		UserID:     user.ID,
		IsAdmin:    user.IsAdmin,
		IsVerified: user.IsVerified,
		ExpiredAt:  accessExpiredAt,
	}).MapClaims(false))

	signedAccessToken, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, errors.New(errors.SigningJWTFailure).Wrap(err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, (model.Auth{
		UserID:    user.ID,
		ExpiredAt: refreshExpiredAt,
	}).MapClaims(true))

	signedRefreshToken, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return nil, errors.New(errors.SigningJWTFailure).Wrap(err)
	}

	a := model.Auth{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
		ExpiredAt:    accessExpiredAt,
	}

	return &a, nil
}

func GenerateAccessToken(user *model.User, tokenExpiry time.Duration, secret string) (*model.Auth, *errors.Error) {
	if user == nil {
		return nil, errors.New(errors.UserNotFound)
	}

	accessExpiredAt := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	if tokenExpiry > 0 {
		accessExpiredAt = time.Now().Add(time.Hour * tokenExpiry).Unix()
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, (model.Auth{
		UserID:     user.ID,
		IsAdmin:    user.IsAdmin,
		IsVerified: user.IsVerified,
		ExpiredAt:  accessExpiredAt,
	}).MapClaims(false))

	signedAccessToken, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, errors.New(errors.SigningJWTFailure).Wrap(err)
	}

	a := model.Auth{
		AccessToken: signedAccessToken,
		ExpiredAt:   accessExpiredAt,
	}

	return &a, nil
}

func ParseToken(token string, secret string) (*model.Auth, *errors.Error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		method, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok || method != jwt.SigningMethodHS256 {
			return nil, errors.New(errors.InvalidJWTSigningMethod)
		}

		return []byte(secret), nil
	})
	if err != nil {
		return nil, errors.New(errors.ParseTokenFailure).Wrap(err)
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid {
		return nil, errors.New(errors.InvalidToken)
	}

	a := model.Auth{
		IsRefreshToken: func(c map[string]any) bool { isRefreshToken, _ := c["isRefreshToken"].(bool); return isRefreshToken }(claims),
		UserID:         func(c map[string]any) string { userID, _ := c["userID"].(string); return userID }(claims),
		IsAdmin:        func(c map[string]any) model.Flag { isAdmin, _ := c["isAdmin"].(bool); return model.Flag(isAdmin) }(claims),
		IsVerified:     func(c map[string]any) model.Flag { IsVrfd, _ := c["isVerified"].(bool); return model.Flag(IsVrfd) }(claims),
		ExpiredAt:      func(c map[string]any) int64 { expiredAt, _ := c["expiredAt"].(float64); return int64(expiredAt) }(claims),
	}

	return &a, nil
}
