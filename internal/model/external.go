package model

type MSGraphAuth struct {
	AccessToken      string  `json:"access_token"`
	RefreshToken     string  `json:"refresh_token"`
	IDToken          string  `json:"id_token"`
	TokenType        string  `json:"token_type"`
	ExpiresIn        int64   `json:"expires_in"`
	ExpiresAt        int64   `json:"-"`
	ExtExpiresIn     int64   `json:"ext_expires_in"`
	Error            string  `json:"error"`
	ErrorDescription string  `json:"error_description"`
	ErrorCodes       []int64 `json:"error_codes"`
}
