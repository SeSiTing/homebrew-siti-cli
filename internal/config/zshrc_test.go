package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// withFakeHome temporarily sets HOME to a tmpdir and returns it.
// Restored automatically via t.Cleanup.
func withFakeHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", old) })
	return dir
}

func TestReadProviders_BasicDiscovery(t *testing.T) {
	home := withFakeHome(t)
	t.Setenv("ZHIPU_API_KEY", "")
	writeFile(t, filepath.Join(home, ".zshrc"), `
export MINIMAX_BASE_URL="https://api.minimaxi.com/anthropic"
export MINIMAX_API_KEY="sk-test"
export ZHIPU_BASE_URL="https://open.bigmodel.cn/api/anthropic"
export ZHIPU_MODEL="glm-4.6"
export ACME_API_ORIGIN="https://acme.example" # composition helper, not a provider
export ANTHROPIC_BASE_URL="..."  # should be skipped (target itself)
export OPENAI_BASE_URL="..."     # should be skipped (target itself)
export TI_BASE_URL="..."         # should be skipped (client target)
`)

	got, err := ReadProviders()
	if err != nil {
		t.Fatalf("ReadProviders: %v", err)
	}
	if len(got) != 6 {
		t.Fatalf("want 6 built-in providers, got %d: %+v", len(got), got)
	}

	mm, ok := got.Find("MINIMAX")
	if !ok {
		t.Fatal("MINIMAX not found")
	}
	if mm.AuthTokenVar != "MINIMAX_API_KEY" {
		t.Errorf("MINIMAX AuthTokenVar = %q, want MINIMAX_API_KEY", mm.AuthTokenVar)
	}
	if mm.ModelVar != "MINIMAX_MODEL" {
		t.Errorf("MINIMAX ModelVar = %q, want MINIMAX_MODEL", mm.ModelVar)
	}

	zp, _ := got.Find("ZHIPU")
	if zp.AuthTokenVar != "ZHIPU_API_KEY" {
		t.Errorf("ZHIPU AuthTokenVar = %q, want ZHIPU_API_KEY", zp.AuthTokenVar)
	}
	if zp.ModelVar != "ZHIPU_MODEL" {
		t.Errorf("ZHIPU ModelVar = %q, want ZHIPU_MODEL", zp.ModelVar)
	}
}

func TestReadProviders_ZshenvWinsOnDuplicate(t *testing.T) {
	home := withFakeHome(t)
	writeFile(t, filepath.Join(home, ".zshenv"), `
export ALI_BASE_URL="https://from-zshenv.example"
export ALI_API_KEY="from-zshenv-key"
export BAILIAN_OPENAI_BASE_URL="https://dashscope.example/v1"
`)
	writeFile(t, filepath.Join(home, ".zshrc"), `
export ALI_BASE_URL="https://from-zshrc.example"
`)

	got, err := ReadProviders()
	if err != nil {
		t.Fatalf("ReadProviders: %v", err)
	}
	if len(got) != 6 {
		t.Fatalf("want 6 built-in providers, got %d", len(got))
	}
	if got[0].AuthTokenVar != "ALI_API_KEY" {
		t.Errorf("zshenv should have won; AuthTokenVar = %q", got[0].AuthTokenVar)
	}
}

func TestReadProviders_GrokMapping(t *testing.T) {
	home := withFakeHome(t)
	writeFile(t, filepath.Join(home, ".zshenv"), `
export ALI_BASE_URL="https://coding.example/apps/anthropic"
export ALI_API_KEY="sk-test"
export ALI_MODEL="qwen3.8-max"
export ALI_GROK_BASE_URL="https://coding.example/v1"
`)

	got, err := ReadProviders()
	if err != nil {
		t.Fatalf("ReadProviders: %v", err)
	}
	if len(got) != 6 {
		t.Fatalf("want 6 built-in providers, got %d: %+v", len(got), got)
	}
	p := got[0]
	if p.GrokBaseURLVar != "ALI_GROK_BASE_URL" {
		t.Errorf("GrokBaseURLVar = %q", p.GrokBaseURLVar)
	}
	if p.GrokAuthTokenVar != "ALI_API_KEY" {
		t.Errorf("GrokAuthTokenVar = %q", p.GrokAuthTokenVar)
	}
	if p.GrokModelVar != "ALI_MODEL" {
		t.Errorf("GrokModelVar = %q", p.GrokModelVar)
	}
	if !p.SupportsGrok() {
		t.Error("provider should support Grok")
	}
}

func TestReadProviders_BuiltInDefaults(t *testing.T) {
	withFakeHome(t)
	t.Setenv("ALI_BASE_URL", "")
	t.Setenv("ALI_GROK_BASE_URL", "")
	t.Setenv("ALI_OPENAI_BASE_URL", "")
	t.Setenv("ALI_CHAT_COMPLETIONS_BASE_URL", "")
	t.Setenv("ALI_MODEL", "")
	t.Setenv("BAILIAN_MODEL", "")
	t.Setenv("BAILIAN_CODEX_MODEL", "")

	got, err := ReadProviders()
	if err != nil {
		t.Fatal(err)
	}
	ali, ok := got.Find("ali")
	if !ok {
		t.Fatal("ALI built-in missing")
	}
	if ali.BaseURL() != "https://coding.dashscope.aliyuncs.com/apps/anthropic" {
		t.Errorf("BaseURL=%q", ali.BaseURL())
	}
	if ali.GrokBaseURL() != "https://coding.dashscope.aliyuncs.com/v1" {
		t.Errorf("GrokBaseURL=%q", ali.GrokBaseURL())
	}
	if ali.Model() != "qwen3.8-max" || ali.GrokModel() != "qwen3.8-max" {
		t.Errorf("models=%q/%q", ali.Model(), ali.GrokModel())
	}
	if ali.AuthTokenVar != "ALI_API_KEY" || ali.GrokAuthTokenVar != "ALI_API_KEY" {
		t.Errorf("built-in auth vars=%q/%q", ali.AuthTokenVar, ali.GrokAuthTokenVar)
	}
	if ali.SupportsCodex() || ali.CodexUnsupportedReason == "" {
		t.Error("ALI Coding Plan should be explicitly unsupported by Codex")
	}

	bailian, ok := got.Find("bailian")
	if !ok || !bailian.SupportsCodex() {
		t.Fatal("BAILIAN built-in should support Codex")
	}
	if bailian.CodexModel() != "qwen3.8-max" {
		t.Errorf("CodexModel=%q", bailian.CodexModel())
	}

	deepseek, ok := got.Find("deepseek")
	if !ok {
		t.Fatal("DEEPSEEK built-in missing")
	}
	if deepseek.BaseURL() != "https://api.deepseek.com/anthropic" {
		t.Errorf("DeepSeek BaseURL=%q", deepseek.BaseURL())
	}
	if deepseek.AuthTokenVar != "DEEPSEEK_API_KEY" || deepseek.Model() != "deepseek-v4-pro" {
		t.Errorf("DeepSeek auth/model=%q/%q", deepseek.AuthTokenVar, deepseek.Model())
	}
	if deepseek.SupportsCodex() || deepseek.CodexUnsupportedReason == "" {
		t.Error("DeepSeek Anthropic should be explicitly unsupported by Codex")
	}
}

func TestReadProviders_DefaultAuthTokenDoesNotSatisfyCustomProvider(t *testing.T) {
	home := withFakeHome(t)
	t.Setenv("DEFAULT_AUTH_TOKEN", "must-not-leak")
	writeFile(t, filepath.Join(home, ".zshenv"), `
export CUSTOM_BASE_URL="https://custom.example/anthropic"
export DEFAULT_AUTH_TOKEN="must-not-leak"
`)

	providers, err := ReadProviders()
	if err != nil {
		t.Fatal(err)
	}
	custom, ok := providers.Find("custom")
	if !ok {
		t.Fatal("CUSTOM provider missing")
	}
	if custom.AuthTokenVar != "CUSTOM_API_KEY" {
		t.Fatalf("AuthTokenVar=%q, want CUSTOM_API_KEY", custom.AuthTokenVar)
	}
}

func TestReadProviders_AnthropicBaseURLSuffixUsesProviderName(t *testing.T) {
	home := withFakeHome(t)
	writeFile(t, filepath.Join(home, ".zshenv"), `
export ACME_ANTHROPIC_BASE_URL="https://acme.example/anthropic"
export ACME_API_KEY="test-key"
export ACME_MODEL="acme-pro"
`)

	providers, err := ReadProviders()
	if err != nil {
		t.Fatal(err)
	}
	acme, ok := providers.Find("acme")
	if !ok {
		t.Fatal("ACME provider missing")
	}
	if acme.BaseURLVar != "ACME_ANTHROPIC_BASE_URL" || acme.AuthTokenVar != "ACME_API_KEY" {
		t.Fatalf("unexpected mapping: %+v", acme)
	}
}

func TestReadProviders_ChatCompletionsBaseURLFeedsGrok(t *testing.T) {
	home := withFakeHome(t)
	writeFile(t, filepath.Join(home, ".zshenv"), `
export ACME_ANTHROPIC_BASE_URL="https://acme.example/anthropic"
export ACME_CHAT_COMPLETIONS_BASE_URL="https://acme.example/v1"
export ACME_API_KEY="test-key"
export ACME_MODEL="acme-pro"
`)

	providers, err := ReadProviders()
	if err != nil {
		t.Fatal(err)
	}
	acme, ok := providers.Find("acme")
	if !ok {
		t.Fatal("ACME provider missing")
	}
	if acme.GrokBaseURLVar != "ACME_CHAT_COMPLETIONS_BASE_URL" || acme.GrokAuthTokenVar != "ACME_API_KEY" || acme.GrokModelVar != "ACME_MODEL" {
		t.Fatalf("unexpected Grok mapping: %+v", acme)
	}
}

func TestReadProviders_ResponsesBaseURLFeedsCodex(t *testing.T) {
	home := withFakeHome(t)
	t.Setenv("ACME_RESPONSES_BASE_URL", "https://acme.example/v1")
	t.Setenv("ACME_API_KEY", "test-key")
	t.Setenv("ACME_MODEL", "acme-coder")
	writeFile(t, filepath.Join(home, ".zshenv"), `
export ACME_ANTHROPIC_BASE_URL="https://acme.example/anthropic"
export ACME_RESPONSES_BASE_URL="https://acme.example/v1"
export ACME_API_KEY="test-key"
export ACME_MODEL="acme-coder"
`)

	providers, err := ReadProviders()
	if err != nil {
		t.Fatal(err)
	}
	acme, ok := providers.Find("acme")
	if !ok {
		t.Fatal("ACME provider missing")
	}
	if acme.CodexBaseURLVar != "ACME_RESPONSES_BASE_URL" || acme.CodexAuthTokenVar != "ACME_API_KEY" || acme.CodexModelVar != "ACME_MODEL" || !acme.SupportsCodex() {
		t.Fatalf("unexpected Codex mapping: %+v", acme)
	}
}

func TestReadProviders_LegacyBaseURLOverride(t *testing.T) {
	home := withFakeHome(t)
	t.Setenv("ALI_ANTHROPIC_BASE_URL", "")
	writeFile(t, filepath.Join(home, ".zshenv"), `
export ALI_BASE_URL="https://legacy.example/anthropic"
export ALI_API_KEY="test-key"
`)

	providers, err := ReadProviders()
	if err != nil {
		t.Fatal(err)
	}
	ali, _ := providers.Find("ali")
	if ali.BaseURLVar != "ALI_BASE_URL" {
		t.Fatalf("legacy BaseURLVar=%q, want ALI_BASE_URL", ali.BaseURLVar)
	}
}

func TestReadProviders_ExplicitAliCodexRouterOverride(t *testing.T) {
	withFakeHome(t)
	t.Setenv("ALI_CODEX_BASE_URL", "http://127.0.0.1:8787/v1")
	providers, err := ReadProviders()
	if err != nil {
		t.Fatal(err)
	}
	ali, _ := providers.Find("ali")
	if !ali.SupportsCodex() || ali.CodexBaseURL() != "http://127.0.0.1:8787/v1" {
		t.Fatalf("explicit router override not enabled: %+v", ali)
	}
}

func TestReadSkipList(t *testing.T) {
	t.Setenv("SITI_AI_SKIP", "OPENAI,BAILIAN, AZURE ")
	got := ReadSkipList()
	want := []string{"OPENAI", "BAILIAN", "AZURE"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got[%d]=%q, want %q", i, got[i], want[i])
		}
	}
}

func TestProvider_IsSkipped(t *testing.T) {
	p := Provider{Name: "MINIMAX"}
	if p.IsSkipped([]string{"OPENAI", "BAILIAN"}) {
		t.Error("MINIMAX should not be skipped")
	}
	if !p.IsSkipped([]string{"openai", "minimax"}) {
		t.Error("MINIMAX should be skipped (case-insensitive)")
	}
}
