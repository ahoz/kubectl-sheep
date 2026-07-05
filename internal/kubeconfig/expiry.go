package kubeconfig

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type rawKubeconfig struct {
	Users []struct {
		Name string `yaml:"name"`
		User struct {
			Token                  string `yaml:"token"`
			ClientCertificateData  string `yaml:"client-certificate-data"`
			ClientKeyData          string `yaml:"client-key-data"`
		} `yaml:"user"`
	} `yaml:"users"`
}

// TokenExpiryHint returns a human-readable expiry hint if one can be extracted.
func TokenExpiryHint(content string) string {
	var cfg rawKubeconfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return ""
	}

	var latest time.Time
	for _, user := range cfg.Users {
		certData := user.User.ClientCertificateData
		if certData == "" {
			continue
		}
		expiry, err := certExpiry(certData)
		if err != nil {
			continue
		}
		if expiry.After(latest) {
			latest = expiry
		}
	}

	if latest.IsZero() {
		if hasBearerToken(cfg) {
			return "token present (expiry not available)"
		}
		return ""
	}
	return fmt.Sprintf("certificate expires %s", latest.UTC().Format(time.RFC3339))
}

func hasBearerToken(cfg rawKubeconfig) bool {
	for _, user := range cfg.Users {
		if strings.TrimSpace(user.User.Token) != "" {
			return true
		}
	}
	return false
}

func certExpiry(b64 string) (time.Time, error) {
	der, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil {
		return time.Time{}, err
	}
	block, _ := pem.Decode(der)
	if block == nil {
		certs, err := x509.ParseCertificates(der)
		if err != nil || len(certs) == 0 {
			return time.Time{}, fmt.Errorf("no certificate found")
		}
		return certs[0].NotAfter, nil
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}
