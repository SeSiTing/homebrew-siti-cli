package cmd

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/SeSiTing/siti-cli/internal/config"
	"github.com/spf13/cobra"
)

func TestApplySwitchIncludesGrokEnvironment(t *testing.T) {
	t.Setenv("ALI_BASE_URL", "https://coding.example/apps/anthropic")
	t.Setenv("ALI_API_KEY", "test-key")
	t.Setenv("ALI_MODEL", "qwen3.8-max")
	t.Setenv("ALI_GROK_BASE_URL", "https://coding.example/v1")
	buf := &evalBuffer{}
	c := &cobra.Command{}
	c.SetContext(context.WithValue(context.Background(), evalKey{}, buf))

	applyShellSwitch(c, config.Provider{
		Name:             "ALI",
		BaseURLVar:       "ALI_BASE_URL",
		AuthTokenVar:     "ALI_API_KEY",
		ModelVar:         "ALI_MODEL",
		GrokBaseURLVar:   "ALI_GROK_BASE_URL",
		GrokAuthTokenVar: "ALI_API_KEY",
		GrokModelVar:     "ALI_MODEL",
	}, aiClients{claude: true, grok: true})

	want := []string{
		`export ANTHROPIC_BASE_URL="$ALI_BASE_URL";`,
		`export ANTHROPIC_AUTH_TOKEN="$ALI_API_KEY";`,
		`export ANTHROPIC_MODEL="$ALI_MODEL";`,
		`export ANTHROPIC_DEFAULT_SONNET_MODEL="$ALI_MODEL";`,
		`export ANTHROPIC_DEFAULT_OPUS_MODEL="$ALI_MODEL";`,
		`export ANTHROPIC_DEFAULT_HAIKU_MODEL="$ALI_MODEL";`,
		`export ANTHROPIC_REASONING_MODEL="$ALI_MODEL";`,
		`export SITI_GROK_MODEL_ID="siti-ali";`,
	}
	if !reflect.DeepEqual(buf.lines, want) {
		t.Fatalf("eval lines:\n got: %#v\nwant: %#v", buf.lines, want)
	}
}

func TestApplySwitchUsesBuiltInNonSecretDefaults(t *testing.T) {
	t.Setenv("ALI_API_KEY", "test-key")
	t.Setenv("ALI_BASE_URL", "")
	t.Setenv("ALI_ANTHROPIC_BASE_URL", "")
	t.Setenv("ALI_MODEL", "")
	t.Setenv("ALI_GROK_BASE_URL", "")
	t.Setenv("ALI_OPENAI_BASE_URL", "")
	t.Setenv("ALI_CHAT_COMPLETIONS_BASE_URL", "")
	t.Setenv("ALI_GROK_MODEL", "")
	buf := &evalBuffer{}
	c := &cobra.Command{}
	c.SetContext(context.WithValue(context.Background(), evalKey{}, buf))

	providers, err := config.ReadProviders()
	if err != nil {
		t.Fatal(err)
	}
	p, ok := providers.Find("ali")
	if !ok {
		t.Fatal("built-in ali provider missing")
	}
	applyShellSwitch(c, p, aiClients{claude: true, grok: true})

	wants := []string{
		`export ANTHROPIC_BASE_URL="https://coding.dashscope.aliyuncs.com/apps/anthropic";`,
		`export ANTHROPIC_MODEL="qwen3.8-max";`,
		`export SITI_GROK_MODEL_ID="siti-ali";`,
	}
	for _, want := range wants {
		found := false
		for _, line := range buf.lines {
			if line == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing %q in %#v", want, buf.lines)
		}
	}
}

func TestParseAIClients(t *testing.T) {
	defaults := aiClients{claude: true, grok: true}
	got, err := parseAIClients("default", defaults)
	if err != nil || !reflect.DeepEqual(got, defaults) {
		t.Fatalf("default=%+v err=%v", got, err)
	}
	got, err = parseAIClients("codex", defaults)
	if err != nil || !reflect.DeepEqual(got, aiClients{codex: true}) {
		t.Fatalf("codex=%+v err=%v", got, err)
	}
	if _, err := parseAIClients("desktop", defaults); err == nil {
		t.Fatal("want invalid client error")
	}
}

func TestPreflightSwitchReportsMissingKey(t *testing.T) {
	t.Setenv("SITI_WRAPPER", "1")
	t.Setenv("ALI_API_KEY", "")
	p := config.Provider{
		Name:               "ALI",
		BaseURLDefault:     "https://example.com/anthropic",
		AuthTokenVar:       "ALI_API_KEY",
		GrokBaseURLDefault: "https://example.com/v1",
		GrokAuthTokenVar:   "ALI_API_KEY",
		GrokModelVar:       "ALI_MODEL",
		ModelDefault:       "qwen3.8-max",
	}
	err := preflightSwitch(p, aiClients{claude: true, grok: true})
	if err == nil || !strings.Contains(err.Error(), "ALI_API_KEY") || !strings.Contains(err.Error(), "未修改任何客户端配置") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPreflightSwitchRejectsAliCodexProtocol(t *testing.T) {
	p := config.Provider{
		Name:                   "ALI",
		CodexAuthTokenVar:      "ALI_API_KEY",
		CodexModelDefault:      "qwen3.8-max",
		CodexUnsupportedReason: "Coding Plan 不支持 Responses API",
	}
	err := preflightSwitch(p, aiClients{codex: true})
	if err == nil || !strings.Contains(err.Error(), "不支持 Responses API") {
		t.Fatalf("unexpected error: %v", err)
	}
}
