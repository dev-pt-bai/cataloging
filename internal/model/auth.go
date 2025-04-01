package model

import "github.com/golang-jwt/jwt/v5"

type Auth struct {
	Token     string `json:"token"`
	UserID    string `json:"-"`
	IsAdmin   bool   `json:"-"`
	ExpiredAt int64  `json:"expiredAt"`
}

func (a Auth) MapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"userID":    a.UserID,
		"isAdmin":   a.IsAdmin,
		"expiredAt": a.ExpiredAt,
	}
}
