package rancher

import "testing"

func TestTokenCreatePageURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://rancher.example.com", "https://rancher.example.com/dashboard/account/create-key"},
		{"https://rancher.example.com/", "https://rancher.example.com/dashboard/account/create-key"},
		{"https://rancher.example.com/v3", "https://rancher.example.com/dashboard/account/create-key"},
		{"https://rancher.example.com:8443/rancher", "https://rancher.example.com:8443/rancher/dashboard/account/create-key"},
	}

	for _, tt := range tests {
		got, err := TokenCreatePageURL(tt.in)
		if err != nil {
			t.Fatalf("TokenCreatePageURL(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("TokenCreatePageURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
