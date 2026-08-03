package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	legacyGrokManagedStart = "# siti-cli: grok model begin"
	legacyGrokManagedEnd   = "# siti-cli: grok model end"
)

// GrokModelID returns the stable model picker ID for a provider.
func GrokModelID(p Provider) string {
	name := strings.ReplaceAll(p.DisplayName(), "_", "-")
	return "siti-" + name
}

// EnsureGrokProvider installs or refreshes a provider-specific, secret-free model entry.
func EnsureGrokProvider(p Provider) (path string, changed bool, err error) {
	if !p.SupportsGrok() {
		return "", false, fmt.Errorf("服务商 '%s' 未配置 Grok", p.DisplayName())
	}
	dir, err := grokConfigDir()
	if err != nil {
		return "", false, err
	}
	path = filepath.Join(dir, "config.toml")
	data, readErr := os.ReadFile(path)
	if readErr != nil && !os.IsNotExist(readErr) {
		return path, false, readErr
	}
	updated, changed, err := mergeGrokProvider(string(data), p)
	if err != nil || !changed {
		return path, changed, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return path, false, err
	}
	if err := writeGrokConfig(path, []byte(updated)); err != nil {
		return path, false, err
	}
	return path, true, nil
}

// GrokConfigReady reports whether at least one siti provider entry is installed.
func GrokConfigReady() bool {
	dir, err := grokConfigDir()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(filepath.Join(dir, "config.toml"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "# siti-cli: grok model siti-")
}

func mergeGrokProvider(content string, p Provider) (string, bool, error) {
	original := content
	content, _, err := removeGrokBlock(content, legacyGrokManagedStart, legacyGrokManagedEnd)
	if err != nil {
		return "", false, err
	}
	id := GrokModelID(p)
	start := "# siti-cli: grok model " + id + " begin"
	end := "# siti-cli: grok model " + id + " end"
	content, _, err = removeGrokBlock(content, start, end)
	if err != nil {
		return "", false, err
	}
	reModel := regexp.MustCompile(`(?m)^\s*\[model\.(?:"` + regexp.QuoteMeta(id) + `"|` + regexp.QuoteMeta(id) + `)\]\s*$`)
	if reModel.MatchString(content) {
		return "", false, fmt.Errorf("Grok 配置已存在 model.%s，请先重命名或删除", id)
	}

	block := strings.Join([]string{
		start,
		"[model." + strconv.Quote(id) + "]",
		"model = " + strconv.Quote(p.GrokModel()),
		"base_url = " + strconv.Quote(p.GrokBaseURL()),
		"name = " + strconv.Quote(p.DisplayName()+" via siti-cli"),
		"env_key = " + strconv.Quote(p.GrokAuthTokenVar),
		`api_backend = "chat_completions"`,
		end,
	}, "\n")
	trimmed := strings.Trim(content, "\n")
	updated := block + "\n"
	if trimmed != "" {
		updated = trimmed + "\n\n" + block + "\n"
	}
	return updated, updated != original, nil
}

func removeGrokBlock(content, startMarker, endMarker string) (string, bool, error) {
	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)
	if (start >= 0) != (end >= 0) || (start >= 0 && end < start) {
		return "", false, fmt.Errorf("Grok 配置中的 siti 管理区块不完整，请手动检查")
	}
	if start < 0 {
		return content, false, nil
	}
	end += len(endMarker)
	return content[:start] + content[end:], true, nil
}

func grokConfigDir() (string, error) {
	if dir := os.Getenv("GROK_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".grok"), nil
}

func writeGrokConfig(path string, data []byte) error {
	dir := filepath.Dir(path)
	mode := os.FileMode(0o600)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	tmp, err := os.CreateTemp(dir, "config.toml.*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
