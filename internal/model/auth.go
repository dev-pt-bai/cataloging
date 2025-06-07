package model

import (
	"errors"
	"fmt"
	"strings"
)

type Auth struct {
	AccessToken    string `json:"accessToken"`
	RefreshToken   string `json:"refreshToken,omitempty"`
	ExpiredAt      int64  `json:"expiredAt"`
	IsRefreshToken bool   `json:"-"`
	UserID         string `json:"-"`
	UserEmail      string `json:"-"`
	Role           Role   `json:"-"`
	IsVerified     Flag   `json:"-"`
}

func (a Auth) IsAdmin() bool {
	return a.Role == Administrator
}

func (a Auth) MapClaims(isRefreshToken bool) map[string]any {
	m := map[string]any{
		"userID":    a.UserID,
		"userEmail": a.UserEmail,
		"expiredAt": a.ExpiredAt,
	}

	if !isRefreshToken {
		m["role"] = a.Role
		m["isVerified"] = a.IsVerified
		return m
	}
	m["isRefreshToken"] = true

	return m
}

type GetTokenRequest struct {
	GrantType    string `jason:"grantType"`
	ID           string `json:"id"`
	Password     string `json:"password"`
	RefreshToken string `json:"refreshToken"`
}

func (r *GetTokenRequest) ValidateLogin() error {
	if r == nil {
		return fmt.Errorf("missing request object")
	}

	messages := make([]string, 0, 5)

	if !strings.EqualFold(r.GrantType, "password") {
		messages = append(messages, fmt.Sprintf("invalid grant type to get token: %s", r.GrantType))
	}

	if len(r.ID) == 0 {
		messages = append(messages, "ID is required")
	}

	if len(r.Password) == 0 {
		messages = append(messages, "password is required")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ","))
	}

	return nil
}

func (r *GetTokenRequest) ValidateRefreshToken() error {
	if r == nil {
		return fmt.Errorf("missing request object")
	}

	messages := make([]string, 0, 5)

	if !strings.EqualFold(r.GrantType, "refreshToken") {
		messages = append(messages, fmt.Sprintf("invalid grant type to get token: %s", r.GrantType))
	}

	if len(r.ID) == 0 {
		messages = append(messages, "ID is required")
	}

	if len(r.RefreshToken) == 0 {
		messages = append(messages, "refresh token is required")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ","))
	}

	return nil
}

type RefreshTokenRequest struct {
	ID string `json:"id"`
}
