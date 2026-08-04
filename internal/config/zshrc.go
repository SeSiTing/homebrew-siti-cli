// Package config provides read-only parsing of shell config files to discover
// AI provider definitions (e.g. MINIMAX_BASE_URL, MINIMAX_API_KEY).
//
// siti never writes to ~/.zshrc or ~/.zshenv; those files are owned by the user.
package config

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Provider represents an AI service provider discovered from shell config files.
type Provider struct {
	// Name is the uppercase prefix, e.g. "MINIMAX", "KIMI", "ALI".
	Name string
	// BaseURLVar is the env var name holding the base URL, e.g. "MINIMAX_BASE_URL".
	BaseURLVar string
	// BaseURLDefault is used when BaseURLVar is not set in the current process.
	BaseURLDefault string
	// AuthTokenVar is the provider-scoped env var used for auth.
	AuthTokenVar string
	// ModelVar is the env var for model overrides, e.g. "MINIMAX_MODEL". Empty if unset.
	ModelVar string
	// ModelDefault is used when ModelVar is not set.
	ModelDefault string
	// GrokBaseURLVar is the optional OpenAI-compatible endpoint used by Grok Build.
	GrokBaseURLVar string
	// GrokBaseURLDefault is the built-in OpenAI-compatible endpoint for Grok Build.
	GrokBaseURLDefault string
	// GrokAuthTokenVar is the env var Grok Build uses for auth.
	GrokAuthTokenVar string
	// GrokModelVar is the model env var used by Grok Build.
	GrokModelVar string
	// CodexBaseURLVar optionally overrides the Responses-compatible Codex endpoint.
	CodexBaseURLVar string
	// CodexBaseURLDefault is the built-in Responses-compatible Codex endpoint.
	CodexBaseURLDefault string
	// CodexAuthTokenVar identifies the key that can be imported into secure storage.
	CodexAuthTokenVar string
	// CodexModelVar optionally overrides the model used by Codex.
	CodexModelVar string
	// CodexModelDefault is the built-in Codex model.
	CodexModelDefault string
	// CodexUnsupportedReason explains why a provider cannot be used by Codex.
	CodexUnsupportedReason string
}

// DisplayName returns the lowercase display name, e.g. "minimax".
func (p Provider) DisplayName() string {
	return strings.ToLower(p.Name)
}

// BaseURL resolves the Claude-compatible endpoint.
func (p Provider) BaseURL() string {
	return envOrDefault(p.BaseURLVar, p.BaseURLDefault)
}

// Model resolves the provider's Claude model.
func (p Provider) Model() string {
	return envOrDefault(p.ModelVar, p.ModelDefault)
}

// GrokBaseURL resolves the provider's Chat Completions endpoint.
func (p Provider) GrokBaseURL() string {
	return envOrDefault(p.GrokBaseURLVar, p.GrokBaseURLDefault)
}

// GrokModel resolves the provider's Grok model.
func (p Provider) GrokModel() string {
	return envOrDefault(p.GrokModelVar, p.Model())
}

// CodexBaseURL resolves the provider's Responses endpoint.
func (p Provider) CodexBaseURL() string {
	return envOrDefault(p.CodexBaseURLVar, p.CodexBaseURLDefault)
}

// CodexModel resolves the provider's Codex model.
func (p Provider) CodexModel() string {
	return envOrDefault(p.CodexModelVar, p.Model())
}

// SupportsGrok reports whether this provider has a complete Grok Build mapping.
func (p Provider) SupportsGrok() bool {
	return p.GrokBaseURL() != "" && p.GrokAuthTokenVar != "" && p.GrokModel() != ""
}

// SupportsCodex reports whether this provider has a Responses-compatible mapping.
func (p Provider) SupportsCodex() bool {
	return p.CodexBaseURL() != "" && p.CodexAuthTokenVar != "" && p.CodexModel() != "" && p.CodexUnsupportedReason == ""
}

// IsSkipped reports whether this provider is in the skip list.
func (p Provider) IsSkipped(skipList []string) bool {
	upper := strings.ToUpper(p.Name)
	for _, s := range skipList {
		if strings.ToUpper(strings.TrimSpace(s)) == upper {
			return true
		}
	}
	return false
}

// ProviderList is a slice of Provider with lookup helpers.
type ProviderList []Provider

// Find returns the provider matching name (case-insensitive) and whether it was found.
func (pl ProviderList) Find(name string) (Provider, bool) {
	upper := strings.ToUpper(name)
	for _, p := range pl {
		if p.Name == upper {
			return p, true
		}
	}
	return Provider{}, false
}

// reExportBaseURL matches lines like MINIMAX_BASE_URL and DEEPSEEK_ANTHROPIC_BASE_URL.
var reExportBaseURL = regexp.MustCompile(`^export\s+([A-Z0-9_]+)_BASE_URL=`)

const defaultQwenModel = "qwen3.8-max"

var builtInProviders = ProviderList{
	{
		Name:                   "ALI",
		BaseURLVar:             "ALI_ANTHROPIC_BASE_URL",
		BaseURLDefault:         "https://coding.dashscope.aliyuncs.com/apps/anthropic",
		AuthTokenVar:           "ALI_API_KEY",
		ModelVar:               "ALI_MODEL",
		ModelDefault:           defaultQwenModel,
		GrokBaseURLVar:         "ALI_CHAT_COMPLETIONS_BASE_URL",
		GrokBaseURLDefault:     "https://coding.dashscope.aliyuncs.com/v1",
		GrokAuthTokenVar:       "ALI_API_KEY",
		GrokModelVar:           "ALI_MODEL",
		CodexAuthTokenVar:      "ALI_API_KEY",
		CodexModelVar:          "ALI_MODEL",
		CodexModelDefault:      defaultQwenModel,
		CodexUnsupportedReason: "Ali Coding Plan 不支持 Codex 所需的 Responses API",
	},
	{
		Name:                "BAILIAN",
		BaseURLVar:          "BAILIAN_ANTHROPIC_BASE_URL",
		BaseURLDefault:      "https://dashscope.aliyuncs.com/apps/anthropic",
		AuthTokenVar:        "BAILIAN_API_KEY",
		ModelVar:            "BAILIAN_MODEL",
		ModelDefault:        defaultQwenModel,
		GrokBaseURLVar:      "BAILIAN_CHAT_COMPLETIONS_BASE_URL",
		GrokBaseURLDefault:  "https://dashscope.aliyuncs.com/compatible-mode/v1",
		GrokAuthTokenVar:    "BAILIAN_API_KEY",
		GrokModelVar:        "BAILIAN_MODEL",
		CodexBaseURLVar:     "BAILIAN_RESPONSES_BASE_URL",
		CodexBaseURLDefault: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		CodexAuthTokenVar:   "BAILIAN_API_KEY",
		CodexModelVar:       "BAILIAN_MODEL",
		CodexModelDefault:   defaultQwenModel,
	},
	{
		Name:                   "DEEPSEEK",
		BaseURLVar:             "DEEPSEEK_ANTHROPIC_BASE_URL",
		BaseURLDefault:         "https://api.deepseek.com/anthropic",
		AuthTokenVar:           "DEEPSEEK_API_KEY",
		ModelVar:               "DEEPSEEK_MODEL",
		ModelDefault:           "deepseek-v4-pro",
		GrokBaseURLVar:         "DEEPSEEK_CHAT_COMPLETIONS_BASE_URL",
		GrokBaseURLDefault:     "https://api.deepseek.com",
		GrokAuthTokenVar:       "DEEPSEEK_API_KEY",
		GrokModelVar:           "DEEPSEEK_MODEL",
		CodexUnsupportedReason: "DeepSeek Anthropic 端点不支持 Codex 所需的 Responses API",
	},
	{
		Name:                   "MINIMAX",
		BaseURLVar:             "MINIMAX_ANTHROPIC_BASE_URL",
		BaseURLDefault:         "https://api.minimaxi.com/anthropic",
		AuthTokenVar:           "MINIMAX_API_KEY",
		ModelVar:               "MINIMAX_MODEL",
		ModelDefault:           "MiniMax-M2.7-highspeed",
		GrokBaseURLVar:         "MINIMAX_CHAT_COMPLETIONS_BASE_URL",
		GrokBaseURLDefault:     "https://api.minimaxi.com/v1",
		GrokAuthTokenVar:       "MINIMAX_API_KEY",
		GrokModelVar:           "MINIMAX_MODEL",
		CodexUnsupportedReason: "MiniMax 未声明 Codex 所需的 Responses API",
	},
	{
		Name:                   "ZHIPU",
		BaseURLVar:             "ZHIPU_ANTHROPIC_BASE_URL",
		BaseURLDefault:         "https://open.bigmodel.cn/api/anthropic",
		AuthTokenVar:           "ZHIPU_API_KEY",
		ModelVar:               "ZHIPU_MODEL",
		ModelDefault:           "glm-5",
		GrokBaseURLVar:         "ZHIPU_CHAT_COMPLETIONS_BASE_URL",
		GrokBaseURLDefault:     "https://open.bigmodel.cn/api/paas/v4",
		GrokAuthTokenVar:       "ZHIPU_API_KEY",
		GrokModelVar:           "ZHIPU_MODEL",
		CodexUnsupportedReason: "智谱未声明 Codex 所需的 Responses API",
	},
	{
		Name:                "OPENROUTER",
		BaseURLVar:          "OPENROUTER_ANTHROPIC_BASE_URL",
		BaseURLDefault:      "https://openrouter.ai/api",
		AuthTokenVar:        "OPENROUTER_API_KEY",
		ModelVar:            "OPENROUTER_MODEL",
		GrokBaseURLVar:      "OPENROUTER_CHAT_COMPLETIONS_BASE_URL",
		GrokBaseURLDefault:  "https://openrouter.ai/api/v1",
		GrokAuthTokenVar:    "OPENROUTER_API_KEY",
		GrokModelVar:        "OPENROUTER_MODEL",
		CodexBaseURLVar:     "OPENROUTER_RESPONSES_BASE_URL",
		CodexBaseURLDefault: "https://openrouter.ai/api/v1",
		CodexAuthTokenVar:   "OPENROUTER_API_KEY",
		CodexModelVar:       "OPENROUTER_MODEL",
	},
}

// ReadProviders parses ~/.zshenv and ~/.zshrc to discover AI provider definitions.
// It reads both files and deduplicates by provider name (zshenv takes precedence).
//
// A provider is recognized from `<NAME>_BASE_URL` or `<NAME>_ANTHROPIC_BASE_URL`.
// The following optional variables are also probed:
//   - <NAME>_API_KEY       → AuthTokenVar (required; no cross-provider fallback)
//   - <NAME>_MODEL          → ModelVar
//   - <NAME>_CHAT_COMPLETIONS_BASE_URL → GrokBaseURLVar
//   - <NAME>_RESPONSES_BASE_URL        → CodexBaseURLVar
//   - <NAME>_OPENAI_BASE_URL / <NAME>_GROK_BASE_URL / <NAME>_CODEX_BASE_URL → legacy overrides
//   - <NAME>_GROK_API_KEY   → GrokAuthTokenVar (falls back to AuthTokenVar)
//   - <NAME>_GROK_MODEL     → GrokModelVar (falls back to ModelVar)
func ReadProviders() (ProviderList, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Collect lines from both files; zshenv first so it wins on duplicate names.
	var lines []string
	for _, name := range []string{".zshenv", ".zshrc"} {
		ls, _ := readLines(filepath.Join(home, name)) // ignore missing files
		lines = append(lines, ls...)
	}

	// Build a set of all defined variable names for fast probing.
	defined := buildDefinedSet(lines)

	// Collect provider names in order (first occurrence wins).
	seen := map[string]bool{}
	var providers ProviderList

	for _, line := range lines {
		m := reExportBaseURL.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		rawName := m[1]
		name := strings.TrimSuffix(rawName, "_ANTHROPIC")
		if name == "ANTHROPIC" || name == "OPENAI" || name == "TI" || strings.HasSuffix(name, "_GROK") || strings.HasSuffix(name, "_CODEX") || strings.HasSuffix(name, "_OPENAI") || strings.HasSuffix(name, "_CHAT_COMPLETIONS") || strings.HasSuffix(name, "_RESPONSES") || seen[name] {
			continue // skip the target variable itself and duplicates
		}
		seen[name] = true

		p, builtIn := builtInProviders.Find(name)
		if !builtIn {
			p = Provider{Name: name, BaseURLVar: rawName + "_BASE_URL"}
		}
		if variableDefined(name+"_API_KEY", defined) {
			p.AuthTokenVar = name + "_API_KEY"
		} else if !builtIn {
			p.AuthTokenVar = name + "_API_KEY"
		}
		if defined[name+"_MODEL"] && p.ModelVar == "" {
			p.ModelVar = name + "_MODEL"
		}
		providers = append(providers, applyProviderOverrides(p, defined))
	}

	for _, p := range builtInProviders {
		if !seen[p.Name] {
			providers = append(providers, applyProviderOverrides(p, defined))
		}
	}

	return providers, nil
}

// ReadSkipList returns the SITI_AI_SKIP provider names from the environment
// or from shell config files (comma-separated).
func ReadSkipList() []string {
	if v := os.Getenv("SITI_AI_SKIP"); v != "" {
		return splitComma(v)
	}

	home, _ := os.UserHomeDir()
	reSITI := regexp.MustCompile(`^export\s+SITI_AI_SKIP=["']?([^"'\n]+)["']?`)

	for _, name := range []string{".zshenv", ".zshrc"} {
		ls, _ := readLines(filepath.Join(home, name))
		for _, line := range ls {
			if m := reSITI.FindStringSubmatch(line); m != nil {
				return splitComma(m[1])
			}
		}
	}
	return nil
}

// CurrentProvider returns the provider name whose BASE_URL matches
// the current ANTHROPIC_BASE_URL environment variable.
func CurrentProvider(providers ProviderList) string {
	current := os.Getenv("ANTHROPIC_BASE_URL")
	if current == "" {
		return ""
	}
	for _, p := range providers {
		if v := os.Getenv(p.BaseURLVar); v != "" && v == current {
			return p.DisplayName()
		}
	}
	return ""
}

// readLines returns all lines from a file, ignoring errors (e.g. file not found).
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// buildDefinedSet returns a set of variable names that appear as `export VAR=` in lines.
var reExportAny = regexp.MustCompile(`^export\s+([A-Z0-9_]+)=`)

func buildDefinedSet(lines []string) map[string]bool {
	set := map[string]bool{}
	for _, line := range lines {
		if m := reExportAny.FindStringSubmatch(line); m != nil {
			set[m[1]] = true
		}
	}
	return set
}

func splitComma(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func envOrDefault(name, fallback string) string {
	if name != "" {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return fallback
}

func applyProviderOverrides(p Provider, defined map[string]bool) Provider {
	name := p.Name
	legacyBaseURLVar := name + "_BASE_URL"
	if p.BaseURLVar != legacyBaseURLVar && !variableDefined(p.BaseURLVar, defined) && variableDefined(legacyBaseURLVar, defined) {
		p.BaseURLVar = legacyBaseURLVar
	}
	if variableDefined(name+"_CHAT_COMPLETIONS_BASE_URL", defined) {
		p.GrokBaseURLVar = name + "_CHAT_COMPLETIONS_BASE_URL"
		p.GrokAuthTokenVar = p.AuthTokenVar
		p.GrokModelVar = p.ModelVar
	} else if variableDefined(name+"_OPENAI_BASE_URL", defined) {
		p.GrokBaseURLVar = name + "_OPENAI_BASE_URL"
		p.GrokAuthTokenVar = p.AuthTokenVar
		p.GrokModelVar = p.ModelVar
	}
	if variableDefined(name+"_GROK_BASE_URL", defined) {
		p.GrokBaseURLVar = name + "_GROK_BASE_URL"
		p.GrokAuthTokenVar = p.AuthTokenVar
		p.GrokModelVar = p.ModelVar
	}
	if variableDefined(name+"_GROK_API_KEY", defined) {
		p.GrokAuthTokenVar = name + "_GROK_API_KEY"
	}
	if variableDefined(name+"_GROK_MODEL", defined) {
		p.GrokModelVar = name + "_GROK_MODEL"
	}
	if variableDefined(name+"_RESPONSES_BASE_URL", defined) {
		p.CodexBaseURLVar = name + "_RESPONSES_BASE_URL"
		p.CodexAuthTokenVar = p.AuthTokenVar
		p.CodexModelVar = p.ModelVar
		p.CodexUnsupportedReason = ""
	} else if variableDefined(name+"_CODEX_BASE_URL", defined) {
		p.CodexBaseURLVar = name + "_CODEX_BASE_URL"
		p.CodexAuthTokenVar = p.AuthTokenVar
		p.CodexModelVar = p.ModelVar
		p.CodexUnsupportedReason = ""
	}
	if variableDefined(name+"_CODEX_API_KEY", defined) {
		p.CodexAuthTokenVar = name + "_CODEX_API_KEY"
	}
	if variableDefined(name+"_CODEX_MODEL", defined) {
		p.CodexModelVar = name + "_CODEX_MODEL"
	}
	return p
}

func variableDefined(name string, defined map[string]bool) bool {
	if defined[name] {
		return true
	}
	value, ok := os.LookupEnv(name)
	return ok && value != ""
}
