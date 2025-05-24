package auth

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
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

	accessToken, err := generateHS256JWT((model.Auth{
		UserID:     user.ID,
		UserEmail:  user.Email,
		IsAdmin:    user.IsAdmin,
		IsVerified: user.IsVerified,
		ExpiredAt:  accessExpiredAt,
	}).MapClaims(false), secret)
	if err != nil {
		return nil, errors.New(errors.GenerateJWTFailure).Wrap(err)
	}

	refreshToken, err := generateHS256JWT((model.Auth{
		UserID:    user.ID,
		ExpiredAt: refreshExpiredAt,
	}).MapClaims(true), secret)
	if err != nil {
		return nil, errors.New(errors.GenerateJWTFailure).Wrap(err)
	}

	a := model.Auth{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
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

	accessToken, err := generateHS256JWT((model.Auth{
		UserID:     user.ID,
		UserEmail:  user.Email,
		IsAdmin:    user.IsAdmin,
		IsVerified: user.IsVerified,
		ExpiredAt:  accessExpiredAt,
	}).MapClaims(false), secret)
	if err != nil {
		return nil, errors.New(errors.GenerateJWTFailure).Wrap(err)
	}

	a := model.Auth{
		AccessToken: accessToken,
		ExpiredAt:   accessExpiredAt,
	}

	return &a, nil
}

func ParseToken(token string, secret string) (*model.Auth, *errors.Error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New(errors.InvalidToken)
	}

	decodedHeader, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New(errors.ParseTokenFailure).Wrap(err)
	}

	header := make(map[string]string)
	if err = json.Unmarshal(decodedHeader, &header); err != nil {
		return nil, errors.New(errors.ParseTokenFailure).Wrap(err)
	}

	if header["typ"] != "JWT" {
		return nil, errors.New(errors.InvalidToken)
	}

	if header["alg"] != "HS256" {
		return nil, errors.New(errors.InvalidJWTSigningMethod)
	}

	expectedSignature := generateSignatureHS256([]byte(parts[0]), []byte(parts[1]), secret)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return nil, errors.New(errors.InvalidToken)
	}

	decodedPayload, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New(errors.ParseTokenFailure).Wrap(err)
	}

	payload := make(map[string]any)
	if err = json.Unmarshal(decodedPayload, &payload); err != nil {
		return nil, errors.New(errors.ParseTokenFailure).Wrap(err)
	}

	a := model.Auth{
		IsRefreshToken: func(c map[string]any) bool { isRefreshToken, _ := c["isRefreshToken"].(bool); return isRefreshToken }(payload),
		UserID:         func(c map[string]any) string { userID, _ := c["userID"].(string); return userID }(payload),
		UserEmail:      func(c map[string]any) string { userEmail, _ := c["userEmail"].(string); return userEmail }(payload),
		IsAdmin:        func(c map[string]any) model.Flag { isAdmin, _ := c["isAdmin"].(bool); return model.Flag(isAdmin) }(payload),
		IsVerified:     func(c map[string]any) model.Flag { IsVrfd, _ := c["isVerified"].(bool); return model.Flag(IsVrfd) }(payload),
		ExpiredAt:      func(c map[string]any) int64 { expiredAt, _ := c["expiredAt"].(float64); return int64(expiredAt) }(payload),
	}

	return &a, nil
}

func GenerateTokenMSGraph(clientID string, aud string, x5t string, key *rsa.PrivateKey) (string, error) {
	header, _ := json.Marshal(map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"x5t": x5t,
	})
	encodedHeader := make([]byte, base64.RawStdEncoding.EncodedLen(len(header)))
	base64.RawURLEncoding.Encode(encodedHeader, header)

	now := time.Now().Unix()
	payload, err := json.Marshal(map[string]any{
		"aud": aud,
		"iss": clientID,
		"sub": clientID,
		"jti": model.NewUUID().String(),
		"nbf": now,
		"exp": now + 300,
	})
	if err != nil {
		return "", err
	}
	encodedPayload := make([]byte, base64.RawStdEncoding.EncodedLen(len(payload)))
	base64.RawURLEncoding.Encode(encodedPayload, payload)

	signature, err := generateSignatureRS256(encodedHeader, encodedPayload, key)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s.%s", encodedHeader, encodedPayload, signature), nil
}

func generateHS256JWT(p map[string]any, secret string) (string, error) {
	if p == nil {
		return "", fmt.Errorf("empty payload")
	}

	if len(secret) == 0 {
		return "", fmt.Errorf("missing hash key")
	}

	header, _ := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	encodedHeader := make([]byte, base64.RawStdEncoding.EncodedLen(len(header)))
	base64.RawURLEncoding.Encode(encodedHeader, header)

	payload, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	encodedPayload := make([]byte, base64.RawStdEncoding.EncodedLen(len(payload)))
	base64.RawURLEncoding.Encode(encodedPayload, payload)

	signature := generateSignatureHS256(encodedHeader, encodedPayload, secret)

	return fmt.Sprintf("%s.%s.%s", encodedHeader, encodedPayload, signature), nil
}

func generateSignatureHS256(header []byte, payload []byte, secret string) string {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(fmt.Appendf(nil, "%s.%s", header, payload))

	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}

func generateSignatureRS256(header []byte, payload []byte, key *rsa.PrivateKey) (string, error) {
	hashed := sha256.Sum256(fmt.Appendf(nil, "%s.%s", header, payload))
	s, err := rsa.SignPKCS1v15(nil, key, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(s), nil
}
