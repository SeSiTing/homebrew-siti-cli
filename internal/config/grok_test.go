package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGrokProviderPreservesExistingContentAndIsIdempotent(t *testing.T) {
	home := withFakeHome(t)
	t.Setenv("GROK_HOME", "")
	dir := filepath.Join(home, ".grok")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.toml")
	writeFile(t, path, "[ui]\ntheme = \"auto\"\n")
	p := Provider{
		Name:               "ALI",
		GrokBaseURLDefault: "https://coding.example/v1",
		GrokAuthTokenVar:   "ALI_API_KEY",
		GrokModelVar:       "ALI_MODEL",
		ModelDefault:       "qwen3.8-max",
	}
	t.Setenv("ALI_MODEL", "")

	gotPath, changed, err := EnsureGrokProvider(p)
	if err != nil {
		t.Fatalf("EnsureGrokProvider: %v", err)
	}
	if gotPath != path || !changed {
		t.Fatalf("path=%q changed=%v", gotPath, changed)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{
		"[ui]",
		"# siti-cli: grok model siti-ali begin",
		`[model."siti-ali"]`,
		`model = "qwen3.8-max"`,
		`base_url = "https://coding.example/v1"`,
		`env_key = "ALI_API_KEY"`,
		`api_backend = "chat_completions"`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("config missing %q:\n%s", want, content)
		}
	}

	beforeSecond, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, changed, err = EnsureGrokProvider(p)
	if err != nil {
		t.Fatalf("second EnsureGrokProvider: %v", err)
	}
	if changed {
		afterSecond, _ := os.ReadFile(path)
		t.Errorf("second EnsureGrokProvider should be idempotent\nbefore:\n%s\nafter:\n%s", beforeSecond, afterSecond)
	}
	if !GrokConfigReady() {
		t.Error("GrokConfigReady = false")
	}
}

func TestEnsureGrokProviderKeepsOtherManagedProviders(t *testing.T) {
	home := withFakeHome(t)
	t.Setenv("GROK_HOME", "")
	dir := filepath.Join(home, ".grok")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	providers := []Provider{
		{Name: "ALI", GrokBaseURLDefault: "https://ali.example/v1", GrokAuthTokenVar: "ALI_API_KEY", GrokModelVar: "ALI_MODEL", ModelDefault: "qwen3.8-max"},
		{Name: "BAILIAN", GrokBaseURLDefault: "https://bailian.example/v1", GrokAuthTokenVar: "BAILIAN_API_KEY", GrokModelVar: "BAILIAN_MODEL", ModelDefault: "qwen3.8-max"},
	}
	for _, p := range providers {
		if _, _, err := EnsureGrokProvider(p); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `[model."siti-ali"]`) || !strings.Contains(content, `[model."siti-bailian"]`) {
		t.Fatalf("managed providers missing:\n%s", content)
	}
}

func TestEnsureGrokProviderRejectsUnmanagedCollision(t *testing.T) {
	home := withFakeHome(t)
	t.Setenv("GROK_HOME", "")
	dir := filepath.Join(home, ".grok")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "config.toml"), "[model.\"siti-ali\"]\nmodel = \"custom\"\n")
	p := Provider{Name: "ALI", GrokBaseURLDefault: "https://example/v1", GrokAuthTokenVar: "ALI_API_KEY", GrokModelVar: "ALI_MODEL", ModelDefault: "qwen3.8-max"}
	if _, _, err := EnsureGrokProvider(p); err == nil {
		t.Fatal("want collision error")
	}
}

func TestEnsureGrokProviderRespectsGrokHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GROK_HOME", dir)
	p := Provider{Name: "ALI", GrokBaseURLDefault: "https://example/v1", GrokAuthTokenVar: "ALI_API_KEY", GrokModelVar: "ALI_MODEL", ModelDefault: "qwen3.8-max"}
	path, changed, err := EnsureGrokProvider(p)
	if err != nil {
		t.Fatalf("EnsureGrokProvider: %v", err)
	}
	if path != filepath.Join(dir, "config.toml") || !changed {
		t.Fatalf("path=%q changed=%v", path, changed)
	}
}
