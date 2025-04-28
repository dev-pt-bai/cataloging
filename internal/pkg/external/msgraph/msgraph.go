package msgraph

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Client struct {
	mu                sync.Mutex
	clientID          string
	scope             string
	redirectURI       string
	encodedThumbprint string
	privateKey        *rsa.PrivateKey
	urlGetToken       string
	urlSendEmail      string
	urlUploadFile     string
	urlDeleteFile     string
	token             *model.MSGraphAuth
	client            *http.Client
}

var client *Client

func NewClient(config *configs.Config) (*Client, error) {
	c := new(Client)
	c.client = http.DefaultClient
	c.urlSendEmail = "https://graph.microsoft.com/v1.0/me/sendMail"

	if config == nil {
		return nil, fmt.Errorf("missing config")
	}

	if len(config.External.MsGraph.ClientID) == 0 {
		return nil, fmt.Errorf("missing msgraph client ID")
	}
	c.clientID = config.External.MsGraph.ClientID

	if len(config.External.MsGraph.Scope) == 0 {
		return nil, fmt.Errorf("missing msgraph scope")
	}
	c.scope = config.External.MsGraph.Scope

	if len(config.External.MsGraph.RedirectURI) == 0 {
		return nil, fmt.Errorf("missing msgraph redirect URI")
	}
	c.redirectURI = config.External.MsGraph.RedirectURI

	if len(config.External.MsGraph.EncodedThumbprint) == 0 {
		return nil, fmt.Errorf("missing msgraph encoded thumbprint")
	}
	c.encodedThumbprint = config.External.MsGraph.EncodedThumbprint

	if config.External.MsGraph.PrivateKey == nil {
		return nil, fmt.Errorf("missing msgraph private key")
	}
	c.privateKey = config.External.MsGraph.PrivateKey

	if len(config.External.MsGraph.TenantID) == 0 {
		return nil, fmt.Errorf("missing msgraph tenant ID")
	}
	c.urlGetToken = fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", config.External.MsGraph.TenantID)

	if len(config.External.MsGraph.DirectoryName) == 0 {
		return nil, fmt.Errorf("missing msgraph directory name")
	}
	c.urlUploadFile = fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/root:/%s", config.External.MsGraph.DirectoryName)

	if len(config.External.MsGraph.DriveID) == 0 {
		return nil, fmt.Errorf("missing msgraph drive ID")
	}
	c.urlDeleteFile = fmt.Sprintf("https://graph.microsoft.com/v1.0/drives/%s/items", config.External.MsGraph.DriveID)

	client = c

	return client, nil
}

func AutoRefreshToken(config *configs.Config) *errors.Error {
	if client == nil || client.token == nil || len(client.token.AccessToken) == 0 {
		slog.Warn("skipping refresh token: please peform delegated authentication first")
		return nil
	}
	slog.Info("refreshing ms graph token")

	return client.refreshToken(context.Background())
}

func (c *Client) refreshToken(ctx context.Context) *errors.Error {
	body, err := c.buildRefreshTokenBody()
	if err != nil {
		return errors.New(errors.MissingMSGraphParameter).Wrap(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.urlGetToken, body)
	if err != nil {
		return errors.New(errors.CreateHTTPRequestFailure).Wrap(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.client.Do(req)
	if err != nil {
		return errors.New(errors.SendHTTPRequestFailure).Wrap(err)
	}
	defer res.Body.Close()

	token := new(model.MSGraphAuth)
	if err := json.NewDecoder(res.Body).Decode(token); err != nil {
		return errors.New(errors.JSONEncodeFailure).Wrap(err)
	}

	if len(token.Error) != 0 {
		err := fmt.Errorf("%s: %s, error codes: %v", token.Error, token.ErrorDescription, token.ErrorCodes)
		return errors.New(errors.GetMSGraphTokenFailure).Wrap(err)
	}

	c.mu.Lock()
	c.token = token
	c.token.ExpiresAt = time.Now().Unix() + token.ExpiresIn
	c.mu.Unlock()

	return nil
}

func (c *Client) buildRefreshTokenBody() (io.Reader, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_id", c.clientID)
	data.Set("scope", c.scope)

	client_assertion, err := c.generateClientAssertion()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client assertion: %w", err)
	}
	data.Set("client_assertion", client_assertion)

	return bytes.NewBufferString(data.Encode()), nil
}

func (c *Client) GetTokenFromAuthCode(ctx context.Context, authCode string) *errors.Error {
	if len(authCode) == 0 {
		return errors.New(errors.InvalidMSGraphAuthCode)
	}

	body, err := c.buildGetTokenFromAuthCodeBody(authCode)
	if err != nil {
		return errors.New(errors.MissingMSGraphParameter).Wrap(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.urlGetToken, body)
	if err != nil {
		return errors.New(errors.CreateHTTPRequestFailure).Wrap(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.client.Do(req)
	if err != nil {
		return errors.New(errors.SendHTTPRequestFailure).Wrap(err)
	}
	defer res.Body.Close()

	token := new(model.MSGraphAuth)
	if err := json.NewDecoder(res.Body).Decode(token); err != nil {
		return errors.New(errors.JSONEncodeFailure).Wrap(err)
	}

	if len(token.Error) != 0 {
		err := fmt.Errorf("%s: %s, error codes: %v", token.Error, token.ErrorDescription, token.ErrorCodes)
		return errors.New(errors.GetMSGraphTokenFailure).Wrap(err)
	}

	c.mu.Lock()
	c.token = token
	c.token.ExpiresAt = time.Now().Unix() + token.ExpiresIn
	c.mu.Unlock()

	return nil
}

func (c *Client) buildGetTokenFromAuthCodeBody(authCode string) (io.Reader, error) {
	data := url.Values{}
	data.Set("code", authCode)
	data.Set("grant_type", "authorization_code")
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_id", c.clientID)
	data.Set("scope", c.scope)
	data.Set("redirect_uri", c.redirectURI)

	client_assertion, err := c.generateClientAssertion()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client assertion: %w", err)
	}
	data.Set("client_assertion", client_assertion)

	return bytes.NewBufferString(data.Encode()), nil
}

func (c *Client) generateClientAssertion() (string, error) {
	now := time.Now().Unix()

	assertion := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud": c.urlGetToken,
		"iss": c.clientID,
		"sub": c.clientID,
		"jti": uuid.NewString(),
		"nbf": now,
		"exp": now + 300,
	})

	assertion.Header["x5t"] = c.encodedThumbprint

	return assertion.SignedString(c.privateKey)
}

func (c *Client) SendEmail(ctx context.Context, email model.Email) *errors.Error {
	if c.token == nil {
		return errors.New(errors.InvalidMSGraphToken)
	}

	if err := email.Validate(); err != nil {
		return errors.New(errors.JSONValidationFailure).Wrap(err)
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(email); err != nil {
		return errors.New(errors.JSONEncodeFailure).Wrap(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.urlSendEmail, body)
	if err != nil {
		return errors.New(errors.CreateHTTPRequestFailure).Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))

	res, err := c.client.Do(req)
	if err != nil {
		return errors.New(errors.SendHTTPRequestFailure).Wrap(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		r := new(model.MSGraphSendEmail)
		if err := json.NewDecoder(res.Body).Decode(r); err != nil {
			return errors.New(errors.JSONDecodeFailure).Wrap(err)
		}
		return errors.New(errors.SendEmailFailure).Wrap(r.Error)
	}

	return nil
}

func (c *Client) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*model.MSGraphUploadFile, *errors.Error) {
	if c.token == nil {
		return nil, errors.New(errors.InvalidMSGraphToken)
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		return nil, errors.New(errors.CreateFormFileFailure).Wrap(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, errors.New(errors.CopyFileFailure).Wrap(err)
	}
	writer.Close()

	ext := filepath.Ext(header.Filename)
	url := fmt.Sprintf("%s/%s%s:/content", c.urlUploadFile, uuid.NewString(), ext)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, errors.New(errors.CreateHTTPRequestFailure).Wrap(err)
	}

	contentType := "application/octet-stream"
	if mimeType := mime.TypeByExtension(ext); len(mimeType) != 0 {
		contentType = mimeType
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))

	res, err := c.client.Do(req)
	if err != nil {
		return nil, errors.New(errors.SendHTTPRequestFailure).Wrap(err)
	}
	defer res.Body.Close()

	f := new(model.MSGraphUploadFile)
	if err := json.NewDecoder(res.Body).Decode(f); err != nil {
		return nil, errors.New(errors.JSONDecodeFailure).Wrap(err)
	}

	if res.StatusCode != http.StatusCreated || f.Error != nil {
		return nil, errors.New(errors.UploadFileFailure).Wrap(f.Error)
	}

	return f, nil
}

func (c *Client) DeleteFile(ctx context.Context, itemID string) *errors.Error {
	if c.token == nil {
		return errors.New(errors.InvalidMSGraphToken)
	}

	url := fmt.Sprintf("%s/%s/permanentDelete", c.urlDeleteFile, itemID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return errors.New(errors.CreateHTTPRequestFailure).Wrap(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))

	res, err := c.client.Do(req)
	if err != nil {
		return errors.New(errors.SendHTTPRequestFailure).Wrap(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		d := new(model.MSGraphDeleteFile)
		if err := json.NewDecoder(res.Body).Decode(d); err != nil {
			return errors.New(errors.JSONDecodeFailure).Wrap(err)
		}

		if d.Error.Code == "itemNotFound" {
			return errors.New(errors.AssetNotFound).Wrap(err)
		}

		return errors.New(errors.DeleteFileFailure).Wrap(d.Error)
	}

	return nil
}
