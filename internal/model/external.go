package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/dev-pt-bai/cataloging/internal/pkg/async/manager"
)

type MSGraphAuth struct {
	AccessToken      string  `json:"access_token"`
	IDToken          string  `json:"id_token"`
	TokenType        string  `json:"token_type"`
	ExpiresIn        int64   `json:"expires_in"`
	ExpiresAt        int64   `json:"-"`
	ExtExpiresIn     int64   `json:"ext_expires_in"`
	Error            string  `json:"error"`
	ErrorDescription string  `json:"error_description"`
	ErrorCodes       []int64 `json:"error_codes"`
}

type MSGraphSendEmail struct {
	Error *MSGraphError `json:"error"`
}

type MSGraphUploadFile struct {
	DownloadURL string        `json:"@microsoft.graph.downloadUrl"`
	WebURL      string        `json:"webUrl"`
	ItemID      string        `json:"id"`
	Name        string        `json:"name"`
	Size        int64         `json:"size"`
	Error       *MSGraphError `json:"error"`
}

func (f *MSGraphUploadFile) Asset(createdBy string) Asset {
	if f == nil {
		return Asset{}
	}

	return Asset{
		ID:          f.ItemID,
		Name:        f.Name,
		Size:        f.Size,
		DownloadURL: f.DownloadURL,
		WebURL:      f.WebURL,
		CreatedBy:   createdBy,
	}
}

type MSGraphDeleteFile struct {
	Error *MSGraphError `json:"error"`
}

type MSGraphError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *MSGraphError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("code: %v, message: %v", e.Code, e.Message)
}

type Body struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type Message struct {
	Subject      string        `json:"subject"`
	Body         Body          `json:"body"`
	ToRecipients []ToRecipient `json:"toRecipients"`
}

type EmailAddress struct {
	Address string `json:"address"`
}

type ToRecipient struct {
	EmailAddress EmailAddress `json:"emailAddress"`
}

type Email struct {
	Message Message `json:"message"`
}

func NewTextEmail(subject, content, recipient string) *Email {
	return &Email{
		Message: Message{
			Subject:      subject,
			Body:         Body{ContentType: "text", Content: content},
			ToRecipients: []ToRecipient{{EmailAddress: EmailAddress{Address: recipient}}},
		},
	}
}

func NewHTMLEmail(subject, content, recipient string) *Email {
	return &Email{
		Message: Message{
			Subject:      subject,
			Body:         Body{ContentType: "html", Content: content},
			ToRecipients: []ToRecipient{{EmailAddress: EmailAddress{Address: recipient}}},
		},
	}
}

func (e *Email) Validate() error {
	if e == nil {
		return errors.New("missing email object")
	}

	messages := make([]string, 0, 5)

	if len(e.Message.Subject) == 0 {
		messages = append(messages, "email subject is required")
	}

	if len(e.Message.Body.Content) == 0 {
		messages = append(messages, "email content is required")
	}

	if len(e.Message.ToRecipients) == 0 || len(e.Message.ToRecipients[0].EmailAddress.Address) == 0 {
		messages = append(messages, "email recipient is required")
	} else {
		if email, err := mail.ParseAddress(e.Message.ToRecipients[0].EmailAddress.Address); err != nil {
			messages = append(messages, fmt.Sprintf("incorrect email format: %s", err.Error()))
		} else {
			at := strings.LastIndex(email.Address, "@")
			if email.Address[at+1:] != "bai.id" {
				messages = append(messages, fmt.Sprintf("incorrect email domain: %s", email.Address[at+1:]))
			}
		}
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}

	return nil
}

func (e *Email) NewTask(taskType string) *manager.Task {
	data, _ := json.Marshal(e)

	return &manager.Task{
		ID:         NewUUID().String(),
		Type:       taskType,
		Data:       data,
		RetryCount: 0,
	}
}
