package model

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

type User struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"-"`
	IsAdmin    Flag   `json:"isAdmin"`
	IsVerified Flag   `json:"isVerified"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

func (u User) NewVerifiedEmail(appBaseURL string) *Email {
	return NewHTMLEmail(
		"[Cataloging] Selamat bergabung",
		fmt.Sprintf(emailWelcome, u.Name, appBaseURL),
		u.Email,
	)
}

type UserOTP struct {
	UserID    string `json:"userID"`
	UserName  string `json:"userName"`
	UserEmail string `json:"userEmail"`
	OTP       string `json:"otp"`
	CreatedAt int64  `json:"createdAt"`
	ExpiredAt int64  `json:"expiredAt"`
}

const src = "123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

func (u User) GenerateOTP() (UserOTP, error) {
	b := make([]byte, 6)
	n, err := io.ReadAtLeast(rand.Reader, b, 6)
	if n < 6 {
		return UserOTP{}, err
	}

	for i := range b {
		b[i] = src[int(b[i])%len(src)]
	}

	return UserOTP{UserID: u.ID, UserName: u.Name, UserEmail: u.Email, OTP: string(b), ExpiredAt: time.Now().Add(1 * time.Hour).Unix()}, nil
}

func (o UserOTP) NewVerificationEmail() *Email {
	expiredAt := time.Unix(o.ExpiredAt, 0).UTC().Add(7 * time.Hour)
	return NewHTMLEmail(
		"[Cataloging] Verifikasi Email Anda",
		fmt.Sprintf(emailVerification, o.UserName, o.OTP, fmt.Sprintf(expiredAt.Format("02 %s 2006 15:04"), indonesianMonth[expiredAt.Month()])),
		o.UserEmail,
	)
}

type Users struct {
	Data  []*User `json:"data"`
	Count int64   `json:"count"`
}

func (u *Users) Scan(src any) error {
	if src == nil {
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert src of type [%T] to []byte", src)
	}

	return json.Unmarshal(b, u)
}

func (u *Users) Response(page Page) map[string]any {
	if u == nil {
		return nil
	}

	return map[string]any{
		"data": u.Data,
		"meta": listMeta(u.Count, page.ItemPerPage, page.Number),
	}
}

type UpsertUserRequest struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *UpsertUserRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("missing request object")
	}

	messages := make([]string, 0, 5)

	if len(r.ID) == 0 {
		messages = append(messages, "user ID is required")
	}

	if len(r.ID) > 250 {
		messages = append(messages, "user ID is too long")
	}

	if len(r.Name) == 0 {
		messages = append(messages, "user name is required")
	}

	if len(r.Name) > 250 {
		messages = append(messages, "user name is too long")
	}

	if len(r.Email) == 0 {
		messages = append(messages, "user email is required")
	} else {
		if email, err := mail.ParseAddress(r.Email); err != nil {
			messages = append(messages, fmt.Sprintf("incorrect email format: %s", err.Error()))
		} else {
			at := strings.LastIndex(email.Address, "@")
			if email.Address[at+1:] != "bai.id" {
				messages = append(messages, fmt.Sprintf("incorrect email domain: %s", email.Address[at+1:]))
			}
		}
	}

	if len(r.Email) > 250 {
		messages = append(messages, "user email is too long")
	}

	if len(r.Password) == 0 {
		messages = append(messages, "user password is required")
	} else {
		if len(r.Password) < 8 {
			messages = append(messages, "password is too short")
		}
		if len(r.Password) > 72 {
			messages = append(messages, "password is too loong")
		}
		if match, _ := regexp.MatchString("[A-Z]", r.Password); !match {
			messages = append(messages, "password must contain uppercase letter(s)")
		}
		if match, _ := regexp.MatchString("[a-z]", r.Password); !match {
			messages = append(messages, "password must contain lowercase letter(s)")
		}
		if match, _ := regexp.MatchString("[0-9]", r.Password); !match {
			messages = append(messages, "password must contain number(s)")
		}
		if match, _ := regexp.MatchString("[^a-zA-Z0-9]", r.Password); !match {
			messages = append(messages, "password must contain special character(s)")
		}
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}

	return nil
}

func (r UpsertUserRequest) Model() User {
	return User{
		ID:       r.ID,
		Name:     r.Name,
		Email:    r.Email,
		Password: r.Password,
	}
}

type VerifyUserRequest struct {
	Code string `json:"code"`
}

func (r *VerifyUserRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("missing request object")
	}

	messages := make([]string, 0, 5)

	if len(r.Code) == 0 {
		messages = append(messages, "verification code is required")
	}

	if len(r.Code) < 6 {
		messages = append(messages, "verification code is too short")
	}

	if len(r.Code) > 6 {
		messages = append(messages, "verification code is too long")
	}

	if match, _ := regexp.MatchString("[^A-Z0-9]", r.Code); match {
		messages = append(messages, "verification code contains illegal characters")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}

	return nil
}

type ListUsersCriteria struct {
	FilterUser
	Sort
	Page
}

type FilterUser struct {
	Name       string
	IsAdmin    *Flag
	IsVerified *Flag
}
