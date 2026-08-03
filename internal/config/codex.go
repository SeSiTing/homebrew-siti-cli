package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	codexManagedProviderID = "siti-managed"
	codexActiveStart       = "# siti-cli: codex active begin"
	codexActiveEnd         = "# siti-cli: codex active end"
	codexProviderStart     = "# siti-cli: codex provider begin"
	codexProviderEnd       = "# siti-cli: codex provider end"
	codexDisabledPrefix    = "# siti-cli: disabled: "
)

var (
	reCodexTopLevel = regexp.MustCompile(`^\s*(model|model_provider)\s*=`)
	reCodexProvider = regexp.MustCompile(`(?m)^\s*\[model_providers\.(?:"siti-managed"|siti-managed)\]\s*$`)
)

// CodexProviderConfig is the global provider selection managed by siti.
type CodexProviderConfig struct {
	ProviderName string
	DisplayName  string
	BaseURL      string
	Model        string
	AuthCommand  string
	AuthArgs     []string
}

// CodexStatus reports the active siti-managed global Codex selection.
type CodexStatus struct {
	Managed      bool
	ProviderName string
	Model        string
	Path         string
}

// ApplyCodexConfig updates only siti's managed Codex blocks and preserves other settings.
func ApplyCodexConfig(cfg CodexProviderConfig) (path, backup string, changed bool, err error) {
	if cfg.ProviderName == "" || cfg.DisplayName == "" || cfg.BaseURL == "" || cfg.Model == "" || cfg.AuthCommand == "" {
		return "", "", false, fmt.Errorf("Codex provider 配置不完整")
	}
	dir, err := codexConfigDir()
	if err != nil {
		return "", "", false, err
	}
	path = filepath.Join(dir, "config.toml")
	data, readErr := os.ReadFile(path)
	if readErr != nil && !os.IsNotExist(readErr) {
		return path, "", false, readErr
	}

	updated, err := mergeCodexConfig(string(data), cfg)
	if err != nil {
		return path, "", false, err
	}
	if updated == string(data) {
		return path, "", false, nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return path, "", false, err
	}
	backup, err = backupCodexConfig(data, readErr == nil)
	if err != nil {
		return path, "", false, err
	}
	if err := writeCodexConfig(path, []byte(updated)); err != nil {
		return path, backup, false, err
	}
	return path, backup, true, nil
}

// ClearCodexConfig restores top-level values disabled by siti and removes managed blocks.
func ClearCodexConfig() (path, backup string, changed bool, err error) {
	dir, err := codexConfigDir()
	if err != nil {
		return "", "", false, err
	}
	path = filepath.Join(dir, "config.toml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return path, "", false, nil
	}
	if err != nil {
		return path, "", false, err
	}
	updated, err := clearCodexConfig(string(data))
	if err != nil {
		return path, "", false, err
	}
	if updated == string(data) {
		return path, "", false, nil
	}
	backup, err = backupCodexConfig(data, true)
	if err != nil {
		return path, "", false, err
	}
	if err := writeCodexConfig(path, []byte(updated)); err != nil {
		return path, backup, false, err
	}
	return path, backup, true, nil
}

// ReadCodexStatus reads siti's managed selection without exposing credentials.
func ReadCodexStatus() (CodexStatus, error) {
	dir, err := codexConfigDir()
	if err != nil {
		return CodexStatus{}, err
	}
	path := filepath.Join(dir, "config.toml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return CodexStatus{Path: path}, nil
	}
	if err != nil {
		return CodexStatus{Path: path}, err
	}
	block, found, err := managedBlock(string(data), codexActiveStart, codexActiveEnd)
	if err != nil || !found {
		return CodexStatus{Path: path}, err
	}
	status := CodexStatus{Managed: true, Path: path}
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# provider: ") {
			status.ProviderName = strings.TrimSpace(strings.TrimPrefix(trimmed, "# provider: "))
		}
		if strings.HasPrefix(trimmed, "model = ") {
			status.Model = unquoteTOML(strings.TrimSpace(strings.TrimPrefix(trimmed, "model = ")))
		}
	}
	return status, nil
}

func mergeCodexConfig(content string, cfg CodexProviderConfig) (string, error) {
	base, _, err := removeManagedBlock(content, codexActiveStart, codexActiveEnd)
	if err != nil {
		return "", err
	}
	base, _, err = removeManagedBlock(base, codexProviderStart, codexProviderEnd)
	if err != nil {
		return "", err
	}
	if reCodexProvider.MatchString(base) {
		return "", fmt.Errorf("Codex 配置已存在 model_providers.%s，请先重命名或删除", codexManagedProviderID)
	}
	base = disableCodexTopLevel(base)

	active := strings.Join([]string{
		codexActiveStart,
		"# provider: " + cfg.ProviderName,
		"model = " + strconv.Quote(cfg.Model),
		"model_provider = " + strconv.Quote(codexManagedProviderID),
		codexActiveEnd,
	}, "\n")

	args := make([]string, 0, len(cfg.AuthArgs))
	for _, arg := range cfg.AuthArgs {
		args = append(args, strconv.Quote(arg))
	}
	provider := strings.Join([]string{
		codexProviderStart,
		"[model_providers." + codexManagedProviderID + "]",
		"name = " + strconv.Quote(cfg.DisplayName),
		"base_url = " + strconv.Quote(cfg.BaseURL),
		`wire_api = "responses"`,
		"",
		"[model_providers." + codexManagedProviderID + ".auth]",
		"command = " + strconv.Quote(cfg.AuthCommand),
		"args = [" + strings.Join(args, ", ") + "]",
		"timeout_ms = 5000",
		"refresh_interval_ms = 0",
		codexProviderEnd,
	}, "\n")

	trimmed := strings.Trim(base, "\n")
	if trimmed == "" {
		return active + "\n\n" + provider + "\n", nil
	}
	return active + "\n\n" + trimmed + "\n\n" + provider + "\n", nil
}

func clearCodexConfig(content string) (string, error) {
	updated, _, err := removeManagedBlock(content, codexActiveStart, codexActiveEnd)
	if err != nil {
		return "", err
	}
	updated, _, err = removeManagedBlock(updated, codexProviderStart, codexProviderEnd)
	if err != nil {
		return "", err
	}
	lines := strings.Split(updated, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, codexDisabledPrefix) {
			lines[i] = strings.TrimPrefix(line, codexDisabledPrefix)
		}
	}
	trimmed := strings.Trim(strings.Join(lines, "\n"), "\n")
	if trimmed == "" {
		return "", nil
	}
	return trimmed + "\n", nil
}

func disableCodexTopLevel(content string) string {
	lines := strings.Split(content, "\n")
	inTable := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			inTable = true
		}
		if !inTable && !strings.HasPrefix(line, codexDisabledPrefix) && reCodexTopLevel.MatchString(line) {
			lines[i] = codexDisabledPrefix + line
		}
	}
	return strings.Join(lines, "\n")
}

func removeManagedBlock(content, startMarker, endMarker string) (string, bool, error) {
	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)
	if (start >= 0) != (end >= 0) || (start >= 0 && end < start) {
		return "", false, fmt.Errorf("Codex 配置中的 siti 管理区块不完整，请手动检查")
	}
	if start < 0 {
		return content, false, nil
	}
	end += len(endMarker)
	return content[:start] + content[end:], true, nil
}

func managedBlock(content, startMarker, endMarker string) (string, bool, error) {
	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)
	if (start >= 0) != (end >= 0) || (start >= 0 && end < start) {
		return "", false, fmt.Errorf("Codex 配置中的 siti 管理区块不完整，请手动检查")
	}
	if start < 0 {
		return "", false, nil
	}
	end += len(endMarker)
	return content[start:end], true, nil
}

func codexConfigDir() (string, error) {
	if dir := os.Getenv("CODEX_HOME"); dir != "" {
		if info, err := os.Stat(dir); err != nil {
			return "", fmt.Errorf("CODEX_HOME 不可用: %w", err)
		} else if !info.IsDir() {
			return "", fmt.Errorf("CODEX_HOME 不是目录: %s", dir)
		}
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex"), nil
}

func backupCodexConfig(data []byte, existed bool) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		stateDir = filepath.Join(home, ".local", "state")
	}
	dir := filepath.Join(stateDir, "siti", "backups", "codex")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	suffix := ".toml"
	if !existed {
		suffix = ".missing"
	}
	path := filepath.Join(dir, "config-"+time.Now().Format("20060102T150405.000000000")+suffix)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func writeCodexConfig(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
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

func unquoteTOML(value string) string {
	unquoted, err := strconv.Unquote(value)
	if err != nil {
		return strings.Trim(value, `"`)
	}
	return unquoted
}
