package model

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	IsAdmin   bool   `json:"isAdmin"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type UpsertUserRequest struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r UpsertUserRequest) Validate() error {
	messages := make([]string, 0, 5)

	if len(r.ID) == 0 {
		messages = append(messages, "user ID is required")
	}

	if len(r.Name) == 0 {
		messages = append(messages, "user name is required")
	}

	if len(r.Email) == 0 {
		messages = append(messages, "user email is required")
	} else {
		if _, err := mail.ParseAddress(r.Email); err != nil {
			messages = append(messages, fmt.Sprintf("incorrect email format: %s", err.Error()))
		}
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
		return errors.New(strings.Join(messages, ","))
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

type LoginRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}
