package auth

import (
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID string, isAdmin bool, config *configs.Config) (string, int64, *errors.Error) {
	expiredAt := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	if config.App.TokenExpiryHour > 0 {
		expiredAt = time.Now().Add(time.Hour * config.App.TokenExpiryHour).Unix()
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, (model.Auth{
		UserID:    userID,
		IsAdmin:   isAdmin,
		ExpiredAt: expiredAt,
	}).MapClaims())

	s, err := t.SignedString([]byte(config.Secret.JWT))
	if err != nil {
		return "", 0, errors.New(errors.SigningJWTFailure).Wrap(err)
	}

	return s, expiredAt, nil
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
		UserID:    func(c map[string]any) string { userID, _ := c["userID"].(string); return userID }(claims),
		IsAdmin:   func(c map[string]any) bool { isAdmin, _ := c["isAdmin"].(bool); return isAdmin }(claims),
		ExpiredAt: func(c map[string]any) int64 { expiredAt, _ := c["expiredAt"].(float64); return int64(expiredAt) }(claims),
	}

	return &a, nil
}
