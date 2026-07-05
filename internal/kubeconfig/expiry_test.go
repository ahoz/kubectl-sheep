package kubeconfig

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func TestTokenExpiryHintFromCertificate(t *testing.T) {
	expiry := time.Date(2026, 12, 31, 12, 0, 0, 0, time.UTC)
	certPEM := mustGenerateCert(t, expiry)
	b64 := base64.StdEncoding.EncodeToString(certPEM)

	content := "users:\n  - name: user\n    user:\n      client-certificate-data: " + b64 + "\n"
	hint := TokenExpiryHint(content)
	if hint == "" {
		t.Fatal("expected expiry hint")
	}
	if want := "certificate expires 2026-12-31T12:00:00Z"; hint != want {
		t.Fatalf("got %q, want %q", hint, want)
	}
}

func TestTokenExpiryHintBearerToken(t *testing.T) {
	content := "users:\n  - name: user\n    user:\n      token: abc123\n"
	hint := TokenExpiryHint(content)
	if hint != "token present (expiry not available)" {
		t.Fatalf("unexpected hint: %q", hint)
	}
}

func mustGenerateCert(t *testing.T, notAfter time.Time) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     notAfter,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, key.Public(), key)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
