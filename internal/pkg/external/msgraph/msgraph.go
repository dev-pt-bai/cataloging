package msgraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Client struct {
	mu           sync.Mutex
	urlGetToken  string
	urlSendEmail string
	token        *model.MSGraphAuth
	config       *configs.Config
	client       *http.Client
}

var client *Client

func NewClient(config *configs.Config) *Client {
	client = &Client{
		urlGetToken:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", config.External.MsGraph.TenantID),
		urlSendEmail: "https://graph.microsoft.com/v1.0/me/sendMail",
		config:       config,
		client:       http.DefaultClient,
	}

	return client
}

func AutoRefreshToken(config *configs.Config) *errors.Error {
	if client == nil || client.token == nil {
		return nil
	}

	return client.getTokenFromRefreshToken(context.Background())
}

func (c *Client) getTokenFromRefreshToken(ctx context.Context) *errors.Error {
	if len(c.token.RefreshToken) == 0 {
		return errors.New(errors.InvalidMSGraphToken)
	}
	slog.InfoContext(ctx, "refreshing ms graph token")

	body, err := c.buildGetTokenFromRefreshTokenBody(c.token.RefreshToken)
	if err != nil {
		return errors.New(errors.MissingMSGraphParameter).Wrap(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.urlGetToken, body)
	if err != nil {
		return errors.New(errors.CreatingHTTPRequestFailure).Wrap(err)
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

func (c *Client) buildGetTokenFromRefreshTokenBody(refreshToken string) (io.Reader, error) {
	data := url.Values{}
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")

	if len(c.config.External.MsGraph.ClientID) == 0 {
		return nil, fmt.Errorf("missing msgraph client ID")
	}
	data.Set("client_id", c.config.External.MsGraph.ClientID)

	client_assertion, err := c.generateClientAssertion()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client assertion: %w", err)
	}
	data.Set("client_assertion", client_assertion)

	if len(c.config.External.MsGraph.Scopes) == 0 {
		return nil, fmt.Errorf("missing msgraph scopes")
	}
	data.Set("scope", strings.Join(c.config.External.MsGraph.Scopes, " "))

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
		return errors.New(errors.CreatingHTTPRequestFailure).Wrap(err)
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

	if len(c.config.External.MsGraph.ClientID) == 0 {
		return nil, fmt.Errorf("missing msgraph client ID")
	}
	data.Set("client_id", c.config.External.MsGraph.ClientID)

	client_assertion, err := c.generateClientAssertion()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client assertion: %w", err)
	}
	data.Set("client_assertion", client_assertion)

	if len(c.config.External.MsGraph.Scopes) == 0 {
		return nil, fmt.Errorf("missing msgraph scopes")
	}
	data.Set("scope", strings.Join(c.config.External.MsGraph.Scopes, " "))

	if len(c.config.External.MsGraph.RedirectURI) == 0 {
		return nil, fmt.Errorf("missing msgraph redirect URI")
	}
	data.Set("redirect_uri", c.config.External.MsGraph.RedirectURI)

	return bytes.NewBufferString(data.Encode()), nil
}

func (c *Client) generateClientAssertion() (string, error) {
	now := time.Now()

	assertion := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud": c.urlGetToken,
		"iss": c.config.External.MsGraph.ClientID,
		"sub": c.config.External.MsGraph.ClientID,
		"jti": uuid.NewString(),
		"nbf": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
	})

	assertion.Header["x5t"] = c.config.External.MsGraph.EncodedThumbprint

	return assertion.SignedString(c.config.External.MsGraph.PrivateKey)
}

func (c *Client) Token(ctx context.Context) (string, *errors.Error) {
	if c.token == nil {
		return "", errors.New(errors.InvalidMSGraphToken)
	}

	if len(c.token.AccessToken) > 0 && c.token.ExpiresAt > time.Now().Unix() {
		return c.token.AccessToken, nil
	}

	if err := c.getTokenFromRefreshToken(ctx); err != nil {
		return "", err
	}

	return c.token.AccessToken, nil
}

func (c *Client) SendEmail(ctx context.Context, email model.Email) *errors.Error {
	if err := email.Validate(); err != nil {
		return errors.New(errors.JSONValidationFailure).Wrap(err)
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(email); err != nil {
		return errors.New(errors.JSONEncodeFailure).Wrap(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.urlSendEmail, body)
	if err != nil {
		return errors.New(errors.CreatingHTTPRequestFailure).Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")

	token, errToken := c.Token(ctx)
	if errToken != nil {
		return errToken
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := c.client.Do(req)
	if err != nil {
		return errors.New(errors.SendHTTPRequestFailure).Wrap(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return errors.New(errors.SendEmailFailure)
	}

	return nil
}
