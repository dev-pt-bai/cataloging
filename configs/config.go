package configs

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"time"
)

type App struct {
	BaseURL     string        `json:"baseURL"`
	Port        uint16        `json:"port"`
	TokenExpiry time.Duration `json:"tokenExpiry"`
}

type Secret struct {
	JWT string `json:"jwt"`
}

type Database struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type External struct {
	MsGraph MsGraph `json:"msGraph"`
}

type MsGraph struct {
	TenantID          string          `json:"tenantID"`
	ClientID          string          `json:"clientID"`
	RedirectURI       string          `json:"redirectURI"`
	Scope             string          `json:"scope"`
	DirectoryName     string          `json:"directoryName"`
	DriveID           string          `json:"driveID"`
	MaxFileSize       int64           `json:"maxFileSize"`
	SupportedFileExt  []string        `json:"supportedFileExt"`
	RefreshInterval   time.Duration   `json:"refreshInterval"`
	PrivateKey        *rsa.PrivateKey `json:"-"`
	EncodedThumbprint string          `json:"-"`
}

type Config struct {
	App      App      `json:"app"`
	Secret   Secret   `json:"secret"`
	Database Database `json:"database"`
	External External `json:"external"`
}

func Load() (*Config, error) {
	f, err := os.Open("./configs/config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open config.json: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read config.json: %w", err)
	}

	config := new(Config)
	if err = json.Unmarshal(b, config); err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}

	f, err = os.Open("./configs/private.key")
	if err != nil {
		return nil, fmt.Errorf("failed to open private.key: %w", err)
	}

	b, err = io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read private.key: %w", err)
	}

	block, _ := pem.Decode(b)
	if block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("unsupported private key format: %s", block.Type)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("unsupported private key type: %T", key)
	}
	config.External.MsGraph.PrivateKey = privateKey

	f, err = os.Open("./configs/certificate.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to open certificate.pem: %w", err)
	}

	b, err = io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate.pem: %w", err)
	}

	block, _ = pem.Decode(b)
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("unsupported private key format: %s", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	thumbprint := sha1.Sum(cert.Raw)
	config.External.MsGraph.EncodedThumbprint = base64.RawURLEncoding.EncodeToString(thumbprint[:])

	return config, nil
}
