package msgraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Client struct {
	urlGetToken string
	token       *model.MSGraphAuth
	config      *configs.Config
	client      *http.Client
}

func NewClient(config *configs.Config) *Client {
	return &Client{
		urlGetToken: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", config.External.MsGraph.TenantID),
		config:      config,
		client:      http.DefaultClient,
	}
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
	c.token = token
	c.token.ExpiresAt = time.Now().Unix() + token.ExpiresIn

	return nil
}

func (c *Client) buildGetTokenFromAuthCodeBody(authCode string) (io.Reader, error) {
	data := url.Values{}
	data.Set("code", authCode)
	data.Set("grant_type", "authorization_code")

	if len(c.config.External.MsGraph.ClientID) == 0 {
		return nil, fmt.Errorf("missing msgraph client ID")
	}
	data.Set("client_id", c.config.External.MsGraph.ClientID)

	if len(c.config.External.MsGraph.ClientSecret) == 0 {
		return nil, fmt.Errorf("missing msgraph client secret")
	}
	data.Set("client_secret", c.config.External.MsGraph.ClientSecret)

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

func (c *Client) GetTokenFromRefreshToken(ctx context.Context, refreshToken string) *errors.Error {
	if len(refreshToken) == 0 {
		return errors.New(errors.InvalidMSGraphRefreshToken)
	}

	body, err := c.buildGetTokenFromRefreshTokenBody(refreshToken)
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
	c.token = token
	c.token.ExpiresAt = time.Now().Unix() + token.ExpiresIn

	return nil
}

func (c *Client) buildGetTokenFromRefreshTokenBody(refreshToken string) (io.Reader, error) {
	data := url.Values{}
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	if len(c.config.External.MsGraph.ClientID) == 0 {
		return nil, fmt.Errorf("missing msgraph client ID")
	}
	data.Set("client_id", c.config.External.MsGraph.ClientID)

	if len(c.config.External.MsGraph.ClientSecret) == 0 {
		return nil, fmt.Errorf("missing msgraph client secret")
	}
	data.Set("client_secret", c.config.External.MsGraph.ClientSecret)

	if len(c.config.External.MsGraph.Scopes) == 0 {
		return nil, fmt.Errorf("missing msgraph scopes")
	}
	data.Set("scope", strings.Join(c.config.External.MsGraph.Scopes, " "))

	return bytes.NewBufferString(data.Encode()), nil
}

func (c *Client) Token(ctx context.Context) (string, *errors.Error) {
	if c.token == nil {
		return "", errors.New(errors.InvalidMSGraphRefreshToken)
	}

	if len(c.token.AccessToken) > 0 && c.token.ExpiresAt > time.Now().Unix() {
		return c.token.AccessToken, nil
	}

	if err := c.GetTokenFromRefreshToken(ctx, c.token.RefreshToken); err != nil {
		return "", err
	}

	return c.token.AccessToken, nil
}
