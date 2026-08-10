package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/SeSiTing/siti-cli/internal/shell"
	"github.com/spf13/cobra"
)

const (
	proxyHost = "127.0.0.1"
	proxyPort = "7890"

	gitProxyConfigPattern = `^(http|https)(\..*)?\.proxy$`
)

type gitProxyEntry struct {
	key   string
	value string
}

type proxyReference struct {
	source   string
	value    string
	endpoint string
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "管理终端代理设置",
}

var proxyOnCmd = &cobra.Command{
	Use:   "on",
	Short: "开启终端代理",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		httpProxy := fmt.Sprintf("http://%s:%s", proxyHost, proxyPort)
		socksProxy := fmt.Sprintf("socks5://%s:%s", proxyHost, proxyPort)
		printErr("✓ 终端代理已开启 (%s:%s)", proxyHost, proxyPort)
		Eval(c,
			shell.Export("http_proxy", httpProxy),
			shell.Export("HTTP_PROXY", httpProxy),
			shell.Export("https_proxy", httpProxy),
			shell.Export("HTTPS_PROXY", httpProxy),
			shell.Export("all_proxy", socksProxy),
			shell.Export("ALL_PROXY", socksProxy),
		)
		return nil
	},
}

var proxyOffCmd = &cobra.Command{
	Use:   "off",
	Short: "关闭终端代理",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		all, err := c.Flags().GetBool("all")
		if err != nil {
			return err
		}

		if all {
			count, err := cleanGlobalGitProxies()
			if err != nil {
				return fmt.Errorf("清理 Git 全局代理: %w", err)
			}
			printErr("✓ 终端代理已关闭")
			printErr("✓ Git 全局代理已清理（%d 项）", count)
		} else {
			printErr("✓ 终端代理已关闭")
		}
		unsetTerminalProxy(c)
		return nil
	},
}

var proxyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前代理状态",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		httpVal := firstNonEmpty(
			lookupEnv("http_proxy"),
			lookupEnv("HTTP_PROXY"),
		)
		httpsVal := firstNonEmpty(
			lookupEnv("https_proxy"),
			lookupEnv("HTTPS_PROXY"),
		)
		allVal := firstNonEmpty(
			lookupEnv("all_proxy"),
			lookupEnv("ALL_PROXY"),
		)

		terminalActive := httpVal != "" || httpsVal != "" || allVal != ""
		fmt.Println("代理状态")
		fmt.Println()
		fmt.Println("终端环境:")
		fmt.Printf("  HTTP:     %s\n", proxyValueOrOff(httpVal))
		fmt.Printf("  HTTPS:    %s\n", proxyValueOrOff(httpsVal))
		fmt.Printf("  SOCKS:    %s\n", proxyValueOrOff(allVal))

		if v := firstNonEmpty(lookupEnv("no_proxy"), lookupEnv("NO_PROXY")); v != "" {
			fmt.Printf("  NO_PROXY: %s\n", v)
		}

		entries, err := globalGitProxies()
		if err != nil {
			return fmt.Errorf("读取 Git 全局代理: %w", err)
		}

		fmt.Println()
		fmt.Println("Git 全局配置:")
		if len(entries) == 0 {
			fmt.Println("  HTTP/HTTPS: off")
		} else {
			for _, entry := range entries {
				fmt.Printf("  %s: %s\n", entry.key, redactProxyURL(entry.value))
			}
		}

		systemActive := false
		systemStatusKnown := runtime.GOOS != "darwin"
		fmt.Println()
		fmt.Println("macOS 系统代理:")
		if runtime.GOOS != "darwin" {
			fmt.Println("  N/A（仅 macOS）")
		} else if proxies, err := macOSSystemProxies(); err != nil {
			fmt.Printf("  ! 检查失败: %v\n", err)
		} else {
			systemStatusKnown = true
			fmt.Printf("  HTTP:  %s\n", proxies["HTTP"])
			fmt.Printf("  HTTPS: %s\n", proxies["HTTPS"])
			fmt.Printf("  SOCKS: %s\n", proxies["SOCKS"])
			systemActive = proxies["HTTP"] != "off" || proxies["HTTPS"] != "off" || proxies["SOCKS"] != "off"
		}

		fmt.Println()
		if !terminalActive && len(entries) == 0 && !systemActive && systemStatusKnown {
			fmt.Println("✓ 未检测到活动代理")
			return nil
		}
		if !systemStatusKnown {
			fmt.Println("! 代理状态未完整确认")
		}
		if terminalActive || len(entries) > 0 {
			fmt.Println("→ 关闭终端和 Git 代理: siti proxy off --all")
		}
		if systemActive {
			fmt.Println("! macOS 系统代理仅展示，不会被 siti proxy off --all 修改")
		}
		return nil
	},
}

func unsetTerminalProxy(c *cobra.Command) {
	Eval(c,
		shell.Unset("http_proxy", "HTTP_PROXY"),
		shell.Unset("https_proxy", "HTTPS_PROXY"),
		shell.Unset("all_proxy", "ALL_PROXY"),
	)
}

func globalGitProxies() ([]gitProxyEntry, error) {
	cmd := exec.Command("git", "config", "--global", "--get-regexp", gitProxyConfigPattern)
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, commandOutputError(err, out)
	}

	var entries []gitProxyEntry
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		entry := gitProxyEntry{key: parts[0]}
		if len(parts) == 2 {
			entry.value = parts[1]
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func cleanGlobalGitProxies() (int, error) {
	entries, err := globalGitProxies()
	if err != nil {
		return 0, err
	}

	keys := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if _, ok := keys[entry.key]; ok {
			continue
		}
		keys[entry.key] = struct{}{}
		cmd := exec.Command("git", "config", "--global", "--unset-all", entry.key)
		if out, err := cmd.CombinedOutput(); err != nil {
			return 0, commandOutputError(err, out)
		}
	}
	return len(entries), nil
}

func commandOutputError(err error, out []byte) error {
	if message := strings.TrimSpace(string(out)); message != "" {
		return fmt.Errorf("%w: %s", err, message)
	}
	return err
}

func proxyValueOrOff(value string) string {
	if value == "" {
		return "off"
	}
	return redactProxyURL(value)
}

func redactProxyURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || parsed.User == nil {
		return value
	}
	parsed.User = url.UserPassword("***", "***")
	return parsed.String()
}

func macOSSystemProxies() (map[string]string, error) {
	out, err := exec.Command("scutil", "--proxy").CombinedOutput()
	if err != nil {
		return nil, commandOutputError(err, out)
	}
	return parseScutilProxies(string(out)), nil
}

func parseScutilProxies(output string) map[string]string {
	values := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), " : ", 2)
		if len(parts) == 2 {
			values[parts[0]] = parts[1]
		}
	}
	return map[string]string{
		"HTTP":  scutilProxyValue(values, "HTTP", "http"),
		"HTTPS": scutilProxyValue(values, "HTTPS", "http"),
		"SOCKS": scutilProxyValue(values, "SOCKS", "socks5"),
	}
}

func scutilProxyValue(values map[string]string, name, scheme string) string {
	if values[name+"Enable"] != "1" {
		return "off"
	}
	host := values[name+"Proxy"]
	port := values[name+"Port"]
	if host == "" {
		return "on"
	}
	if port == "" {
		return host
	}
	return fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(host, port))
}

func preflightUpgradeProxy() error {
	references, err := configuredLocalProxyReferences()
	if err != nil {
		return fmt.Errorf("读取代理配置: %w", err)
	}

	checked := make(map[string]bool)
	var stale []proxyReference
	for _, reference := range references {
		available, ok := checked[reference.endpoint]
		if !ok {
			connection, err := net.DialTimeout("tcp", reference.endpoint, 300*time.Millisecond)
			available = err == nil
			if connection != nil {
				_ = connection.Close()
			}
			checked[reference.endpoint] = available
		}
		if !available {
			stale = append(stale, reference)
		}
	}
	if len(stale) == 0 {
		return nil
	}

	var lines []string
	for _, reference := range stale {
		lines = append(lines, fmt.Sprintf("  %s: %s（%s 没有进程监听）", reference.source, redactProxyURL(reference.value), reference.endpoint))
	}
	return fmt.Errorf("本地代理不可用:\n%s\n\n升级尚未开始，未修改任何代理配置。\n\n→ 清理代理:\n  siti proxy off --all", strings.Join(lines, "\n"))
}

func configuredLocalProxyReferences() ([]proxyReference, error) {
	var references []proxyReference
	terminal := []struct {
		name  string
		value string
	}{
		{name: "终端 HTTP", value: firstNonEmpty(lookupEnv("http_proxy"), lookupEnv("HTTP_PROXY"))},
		{name: "终端 HTTPS", value: firstNonEmpty(lookupEnv("https_proxy"), lookupEnv("HTTPS_PROXY"))},
		{name: "终端 SOCKS", value: firstNonEmpty(lookupEnv("all_proxy"), lookupEnv("ALL_PROXY"))},
	}
	for _, item := range terminal {
		value := item.value
		if endpoint, ok := localProxyEndpoint(value); ok {
			references = append(references, proxyReference{source: item.name, value: value, endpoint: endpoint})
		}
	}

	entries, err := globalGitProxies()
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !gitProxyAppliesToGitHub(entry.key) {
			continue
		}
		if endpoint, ok := localProxyEndpoint(entry.value); ok {
			references = append(references, proxyReference{source: "Git global " + entry.key, value: entry.value, endpoint: endpoint})
		}
	}
	return references, nil
}

func gitProxyAppliesToGitHub(key string) bool {
	key = strings.ToLower(key)
	if key == "http.proxy" || key == "https.proxy" {
		return true
	}
	remainder, ok := strings.CutPrefix(key, "http.https://github.com")
	return ok && (remainder == ".proxy" || strings.HasPrefix(remainder, "/"))
}

func localProxyEndpoint(value string) (string, bool) {
	if value == "" {
		return "", false
	}
	candidate := value
	if !strings.Contains(candidate, "://") {
		candidate = "http://" + candidate
	}
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.Port() == "" {
		return "", false
	}
	host := parsed.Hostname()
	ip := net.ParseIP(host)
	if !strings.EqualFold(host, "localhost") && (ip == nil || !ip.IsLoopback()) {
		return "", false
	}
	return net.JoinHostPort(host, parsed.Port()), true
}

func init() {
	proxyOffCmd.Flags().Bool("all", false, "同时清理 Git 全局 HTTP/HTTPS 代理")
	proxyCmd.AddCommand(proxyOnCmd, proxyOffCmd, proxyStatusCmd)
	rootCmd.AddCommand(proxyCmd)
}
