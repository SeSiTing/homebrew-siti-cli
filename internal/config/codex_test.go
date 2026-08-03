package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyCodexConfigPreservesAndRestoresUserSettings(t *testing.T) {
	home := withFakeHome(t)
	dir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", dir)
	path := filepath.Join(dir, "config.toml")
	original := `model = "gpt-5.6"
model_provider = "openai"
approval_policy = "on-request"

[mcp_servers.docs]
url = "https://example.com/mcp"
`
	writeFile(t, path, original)

	gotPath, backup, changed, err := ApplyCodexConfig(CodexProviderConfig{
		ProviderName: "bailian",
		DisplayName:  "Bailian",
		BaseURL:      "https://dashscope.example/v1",
		Model:        "qwen3.8-max",
		AuthCommand:  "/opt/homebrew/bin/siti",
		AuthArgs:     []string{"ai", "credential-helper", "bailian"},
	})
	if err != nil {
		t.Fatalf("ApplyCodexConfig: %v", err)
	}
	if gotPath != path || backup == "" || !changed {
		t.Fatalf("path=%q backup=%q changed=%v", gotPath, backup, changed)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{
		codexActiveStart,
		`model = "qwen3.8-max"`,
		`model_provider = "siti-managed"`,
		codexDisabledPrefix + `model = "gpt-5.6"`,
		codexDisabledPrefix + `model_provider = "openai"`,
		`approval_policy = "on-request"`,
		`[mcp_servers.docs]`,
		`[model_providers.siti-managed.auth]`,
		`wire_api = "responses"`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("config missing %q:\n%s", want, content)
		}
	}
	if _, _, changed, err := ApplyCodexConfig(CodexProviderConfig{
		ProviderName: "bailian",
		DisplayName:  "Bailian",
		BaseURL:      "https://dashscope.example/v1",
		Model:        "qwen3.8-max",
		AuthCommand:  "/opt/homebrew/bin/siti",
		AuthArgs:     []string{"ai", "credential-helper", "bailian"},
	}); err != nil || changed {
		t.Fatalf("second apply: changed=%v err=%v", changed, err)
	}

	status, err := ReadCodexStatus()
	if err != nil {
		t.Fatal(err)
	}
	if !status.Managed || status.ProviderName != "bailian" || status.Model != "qwen3.8-max" {
		t.Fatalf("status=%+v", status)
	}

	_, clearBackup, changed, err := ClearCodexConfig()
	if err != nil {
		t.Fatalf("ClearCodexConfig: %v", err)
	}
	if clearBackup == "" || !changed {
		t.Fatalf("clear backup=%q changed=%v", clearBackup, changed)
	}
	restored, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != original {
		t.Fatalf("restored config:\n%s\nwant:\n%s", restored, original)
	}
}

func TestApplyCodexConfigRejectsUnmanagedCollision(t *testing.T) {
	home := withFakeHome(t)
	dir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", dir)
	writeFile(t, filepath.Join(dir, "config.toml"), "[model_providers.siti-managed]\nname = \"mine\"\n")

	_, _, _, err := ApplyCodexConfig(CodexProviderConfig{
		ProviderName: "bailian",
		DisplayName:  "Bailian",
		BaseURL:      "https://example.com/v1",
		Model:        "qwen3.8-max",
		AuthCommand:  "/usr/local/bin/siti",
	})
	if err == nil {
		t.Fatal("want collision error")
	}
}

func TestApplyCodexConfigRequiresExistingExplicitCodexHome(t *testing.T) {
	home := withFakeHome(t)
	t.Setenv("CODEX_HOME", filepath.Join(home, "missing"))
	_, _, _, err := ApplyCodexConfig(CodexProviderConfig{
		ProviderName: "bailian",
		DisplayName:  "Bailian",
		BaseURL:      "https://example.com/v1",
		Model:        "qwen3.8-max",
		AuthCommand:  "/usr/local/bin/siti",
	})
	if err == nil {
		t.Fatal("want CODEX_HOME error")
	}
}
