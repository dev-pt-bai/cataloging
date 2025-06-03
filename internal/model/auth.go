package model

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

type LoginRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	ID string `json:"id"`
}
