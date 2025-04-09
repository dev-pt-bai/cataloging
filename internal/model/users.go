package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/mail"
	"regexp"
	"strings"
)

type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	IsAdmin   Flag   `json:"isAdmin"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
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

func (u *Users) Reponse(page Page) map[string]any {
	if u == nil {
		return nil
	}
	totalPages := int64(math.Ceil(float64(u.Count) / float64(page.ItemPerPage)))
	return map[string]any{
		"data": u.Data,
		"meta": map[string]any{
			"totalRecords": u.Count,
			"totalPages":   totalPages,
			"currentPage":  page.Number,
			"previousPage": func(currentPage, totalPages int64) *int64 {
				if currentPage == 1 || currentPage > totalPages+1 {
					return nil
				}
				currentPage--
				return &currentPage
			}(page.Number, totalPages),
			"nextPage": func(currentPage, totalPages int64) *int64 {
				if currentPage >= totalPages {
					return nil
				}
				currentPage++
				return &currentPage
			}(page.Number, totalPages),
		},
	}
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
		if _, err := mail.ParseAddress(r.Email); err != nil {
			messages = append(messages, fmt.Sprintf("incorrect email format: %s", err.Error()))
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

type ListUsersCriteria struct {
	FilterUser
	Sort
	Page
}

type FilterUser struct {
	Name    string
	IsAdmin *Flag
}
