package auth

import (
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID string, isAdmin model.Flag, config *configs.Config) (*model.Auth, *errors.Error) {
	accessExpiredAt := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	refreshExpiredAt := accessExpiredAt
	if config.App.TokenExpiry > 0 {
		now := time.Now()
		accessExpiredAt = now.Add(time.Hour * config.App.TokenExpiry).Unix()
		refreshExpiredAt = now.Add(10 * time.Hour * config.App.TokenExpiry).Unix()
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, (model.Auth{
		UserID:    userID,
		IsAdmin:   isAdmin,
		ExpiredAt: accessExpiredAt,
	}).MapClaims(false))

	signedAccessToken, err := accessToken.SignedString([]byte(config.Secret.JWT))
	if err != nil {
		return nil, errors.New(errors.SigningJWTFailure).Wrap(err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, (model.Auth{
		UserID:    userID,
		ExpiredAt: refreshExpiredAt,
	}).MapClaims(true))

	signedRefreshToken, err := refreshToken.SignedString([]byte(config.Secret.JWT))
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

func GenerateAccessToken(userID string, isAdmin model.Flag, config *configs.Config) (*model.Auth, *errors.Error) {
	accessExpiredAt := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	if config.App.TokenExpiry > 0 {
		accessExpiredAt = time.Now().Add(time.Hour * config.App.TokenExpiry).Unix()
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, (model.Auth{
		UserID:    userID,
		IsAdmin:   isAdmin,
		ExpiredAt: accessExpiredAt,
	}).MapClaims(false))

	signedAccessToken, err := accessToken.SignedString([]byte(config.Secret.JWT))
	if err != nil {
		return nil, errors.New(errors.SigningJWTFailure).Wrap(err)
	}

	a := model.Auth{
		AccessToken: signedAccessToken,
		ExpiredAt:   accessExpiredAt,
	}

	return &a, nil
}

func ParseToken(token string, config *configs.Config) (*model.Auth, *errors.Error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		method, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok || method != jwt.SigningMethodHS256 {
			return nil, errors.New(errors.InvalidJWTSigningMethod)
		}

		return []byte(config.Secret.JWT), nil
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
		ExpiredAt:      func(c map[string]any) int64 { expiredAt, _ := c["expiredAt"].(float64); return int64(expiredAt) }(claims),
	}

	return &a, nil
}
