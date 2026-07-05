package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"https://rancher.example.com", true},
		{"http://localhost:8080", true},
		{"not-a-url", false},
		{"", false},
		{"https://", false},
	}

	for _, tt := range tests {
		err := ValidateURL(tt.input)
		if tt.ok && err != nil {
			t.Errorf("ValidateURL(%q) unexpected error: %v", tt.input, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("ValidateURL(%q) expected error", tt.input)
		}
	}
}

func TestAddRemoveInstance(t *testing.T) {
	cfg := &Config{}

	inst := Instance{
		Name:    "prod",
		URL:     "https://rancher.prod.example.com",
		Storage: StorageEncrypted,
	}
	if err := cfg.AddInstance(inst); err != nil {
		t.Fatalf("AddInstance: %v", err)
	}

	if err := cfg.ValidateName("prod"); err == nil {
		t.Fatal("expected duplicate name error")
	}

	if err := cfg.RemoveInstance("prod"); err != nil {
		t.Fatalf("RemoveInstance: %v", err)
	}
	if len(cfg.Instances) != 0 {
		t.Fatalf("expected empty instances, got %d", len(cfg.Instances))
	}
}

func TestSaveLoad(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := &Config{
		Instances: []Instance{
			{
				Name:               "dev",
				URL:                "https://rancher.dev.example.com",
				InsecureSkipVerify: true,
				Storage:            StoragePlaintext,
			},
		},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path, err := InstancesPath()
	if err != nil {
		t.Fatalf("InstancesPath: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600, got %o", info.Mode().Perm())
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Instances) != 1 || loaded.Instances[0].Name != "dev" {
		t.Fatalf("unexpected loaded config: %+v", loaded.Instances)
	}

	dir := filepath.Join(home, ".config", configDirName)
	info, err = os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("expected dir mode 0700, got %o", info.Mode().Perm())
	}
}
