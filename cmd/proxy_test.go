package cmd

import (
	"context"
	"net"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestUnsetTerminalProxy(t *testing.T) {
	buf := &evalBuffer{}
	c := &cobra.Command{}
	c.SetContext(context.WithValue(context.Background(), evalKey{}, buf))

	unsetTerminalProxy(c)

	want := []string{
		"unset http_proxy HTTP_PROXY;",
		"unset https_proxy HTTPS_PROXY;",
		"unset all_proxy ALL_PROXY;",
	}
	if !reflect.DeepEqual(buf.lines, want) {
		t.Fatalf("eval lines:\n got: %#v\nwant: %#v", buf.lines, want)
	}
}

func TestEnableAndDisableGlobalGitProxy(t *testing.T) {
	configFile := t.TempDir() + "/gitconfig"
	t.Setenv("GIT_CONFIG_GLOBAL", configFile)

	setGitConfig(t, "http.https://github.com.proxy", "http://127.0.0.1:7891")
	setGitConfig(t, "http.sslVerify", "false")

	if err := enableGlobalGitProxy(); err != nil {
		t.Fatal(err)
	}
	entries, err := globalGitProxies()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("globalGitProxies() returned %d entries, want 3: %#v", len(entries), entries)
	}
	for _, key := range []string{"http.proxy", "https.proxy"} {
		out, err := exec.Command("git", "config", "--global", "--get", key).Output()
		if err != nil || string(out) != "http://127.0.0.1:7890\n" {
			t.Fatalf("%s not enabled correctly: value=%q err=%v", key, out, err)
		}
	}

	count, err := disableGlobalGitProxy()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("disableGlobalGitProxy() = %d, want 2", count)
	}

	entries, err = globalGitProxies()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].key != "http.https://github.com.proxy" {
		t.Fatalf("unexpected Git proxies after off: %#v", entries)
	}

	out, err := exec.Command("git", "config", "--global", "--get", "http.sslVerify").Output()
	if err != nil || string(out) != "false\n" {
		t.Fatalf("unrelated Git config changed: value=%q err=%v", out, err)
	}
}

func TestParseScutilProxies(t *testing.T) {
	proxies := parseScutilProxies(`
<dictionary> {
  HTTPEnable : 1
  HTTPProxy : 127.0.0.1
  HTTPPort : 7890
  HTTPSEnable : 0
  SOCKSEnable : 1
  SOCKSProxy : localhost
  SOCKSPort : 7891
}
`)
	want := map[string]string{
		"HTTP":  "http://127.0.0.1:7890",
		"HTTPS": "off",
		"SOCKS": "socks5://localhost:7891",
	}
	if !reflect.DeepEqual(proxies, want) {
		t.Fatalf("parseScutilProxies() = %#v, want %#v", proxies, want)
	}
}

func TestLocalProxyEndpoint(t *testing.T) {
	tests := []struct {
		value string
		want  string
		ok    bool
	}{
		{"http://127.0.0.1:7890", "127.0.0.1:7890", true},
		{"socks5://localhost:7891", "localhost:7891", true},
		{"http://[::1]:7892", "[::1]:7892", true},
		{"http://proxy.example.com:8080", "", false},
		{"http://127.0.0.1", "", false},
	}
	for _, tt := range tests {
		got, ok := localProxyEndpoint(tt.value)
		if got != tt.want || ok != tt.ok {
			t.Errorf("localProxyEndpoint(%q) = %q, %v; want %q, %v", tt.value, got, ok, tt.want, tt.ok)
		}
	}
}

func TestRedactProxyURL(t *testing.T) {
	got := redactProxyURL("http://user:secret@127.0.0.1:7890")
	if strings.Contains(got, "user") || strings.Contains(got, "secret") {
		t.Fatalf("proxy credentials were not redacted: %s", got)
	}
	if !strings.Contains(got, "127.0.0.1:7890") {
		t.Fatalf("proxy endpoint missing after redaction: %s", got)
	}
}

func TestParseClashProcesses(t *testing.T) {
	status := parseClashProcesses(`
/Library/PrivilegedHelperTools/party.mihomo.helper
/Library/PrivilegedHelperTools/io.github.clash-verge-rev.clash-verge-rev.service.bundle/Contents/MacOS/clash-verge-service
/Applications/Clash Verge.app/Contents/MacOS/clash-verge
/Applications/Clash Verge.app/Contents/MacOS/verge-mihomo -d /tmp/verge -f /Users/test/Library/Application Support/io.github.clash-verge-rev.clash-verge-rev/clash-verge.yaml -ext-ctl-unix /tmp/verge/verge-mihomo.sock
`)

	if !status.appRunning || !status.helperRunning || !status.coreRunning {
		t.Fatalf("parseClashProcesses() = %#v", status)
	}
	wantPath := "/Users/test/Library/Application Support/io.github.clash-verge-rev.clash-verge-rev/clash-verge.yaml"
	if status.configPath != wantPath {
		t.Fatalf("config path = %q, want %q", status.configPath, wantPath)
	}
}

func TestParseClashProcessesDoesNotTreatHelpersAsCore(t *testing.T) {
	status := parseClashProcesses(`
/Library/PrivilegedHelperTools/party.mihomo.helper
/Users/test/.config/clash/service/clash-core-service
`)

	if !status.helperRunning || status.appRunning || status.coreRunning {
		t.Fatalf("helper-only status = %#v", status)
	}
}

func TestLoadClashTunConfig(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(path, []byte(`
mixed-port: 7890
tun:
  enable: true
  device: utun1024
`), 0o600); err != nil {
		t.Fatal(err)
	}

	config, known := loadClashTunConfig([]string{path})
	if !known || !config.enabled || config.device != "utun1024" {
		t.Fatalf("loadClashTunConfig() = %#v, %v", config, known)
	}
}

func TestClashSystemProxyConfigured(t *testing.T) {
	if !clashSystemProxyConfigured(map[string]string{
		"HTTP":  "http://127.0.0.1:7890",
		"HTTPS": "off",
		"SOCKS": "off",
	}) {
		t.Fatal("Clash system proxy was not detected")
	}
	if clashSystemProxyConfigured(map[string]string{
		"HTTP":  "http://192.168.1.2:7890",
		"HTTPS": "off",
		"SOCKS": "off",
	}) {
		t.Fatal("remote proxy was incorrectly detected as local Clash")
	}
}

func TestClashStatusLinesHelperOnly(t *testing.T) {
	status := clashRuntimeStatus{
		processKnown:          true,
		helperRunning:         true,
		systemProxyKnown:      true,
		tunConfigKnown:        true,
		tunConfigured:         true,
		tunDevice:             "utun1024",
		portOpen:              false,
		coreRunning:           false,
		appRunning:            false,
		systemProxyConfigured: false,
	}
	got := strings.Join(clashStatusLines(status), "\n")
	for _, want := range []string{
		"主程序:     未运行",
		"后台服务:   运行中（不代表代理已开启）",
		"代理核心:   未运行",
		"系统代理:   关闭",
		"虚拟网卡:   关闭（配置已开启，代理核心未运行）",
		"本地端口:   7890 closed",
		"当前接管:   未接管",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("status output missing %q:\n%s", want, got)
		}
	}
}

func TestClashStatusLinesBothModes(t *testing.T) {
	status := clashRuntimeStatus{
		processKnown:          true,
		appRunning:            true,
		coreRunning:           true,
		portOpen:              true,
		systemProxyKnown:      true,
		systemProxyConfigured: true,
		tunConfigKnown:        true,
		tunConfigured:         true,
		tunDevice:             "utun1024",
		tunInterfaceReady:     true,
		tunActive:             true,
	}
	got := strings.Join(clashStatusLines(status), "\n")
	for _, want := range []string{
		"系统代理:   开启（127.0.0.1:7890）",
		"虚拟网卡:   开启（utun1024）",
		"当前接管:   系统代理 + 虚拟网卡",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("status output missing %q:\n%s", want, got)
		}
	}
}

func TestPreflightUpgradeProxy(t *testing.T) {
	clearProxyEnvironment(t)
	configFile := t.TempDir() + "/gitconfig"
	t.Setenv("GIT_CONFIG_GLOBAL", configFile)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	workingProxy := "http://" + listener.Addr().String()
	setGitConfig(t, "https.proxy", workingProxy)
	if err := preflightUpgradeProxy(); err != nil {
		listener.Close()
		t.Fatalf("working local proxy rejected: %v", err)
	}

	deadAddress := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	setGitConfig(t, "https.proxy", "http://"+deadAddress)

	err = preflightUpgradeProxy()
	if err == nil {
		t.Fatal("dead local proxy was not rejected")
	}
	for _, want := range []string{"Git global https.proxy", deadAddress, "siti proxy git off", "未修改任何代理配置"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("preflight error %q does not contain %q", err, want)
		}
	}
}

func TestPreflightUpgradeProxyIgnoresURLScope(t *testing.T) {
	clearProxyEnvironment(t)
	t.Setenv("GIT_CONFIG_GLOBAL", t.TempDir()+"/gitconfig")
	setGitConfig(t, "http.https://github.com.proxy", "http://127.0.0.1:1")

	if err := preflightUpgradeProxy(); err != nil {
		t.Fatalf("URL-scoped proxy blocked upgrade: %v", err)
	}
}

func TestPreflightUpgradeProxySuggestsTerminalOff(t *testing.T) {
	clearProxyEnvironment(t)
	t.Setenv("GIT_CONFIG_GLOBAL", t.TempDir()+"/gitconfig")

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	deadAddress := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HTTPS_PROXY", "http://"+deadAddress)

	err = preflightUpgradeProxy()
	if err == nil || !strings.Contains(err.Error(), "siti proxy off") {
		t.Fatalf("terminal proxy remediation missing: %v", err)
	}
	if strings.Contains(err.Error(), "siti proxy git off") {
		t.Fatalf("unexpected Git remediation for terminal-only proxy: %v", err)
	}
}

func TestBrewProxyHintWhenOnlyGitProxyIsEnabled(t *testing.T) {
	clearProxyEnvironment(t)
	t.Setenv("GIT_CONFIG_GLOBAL", t.TempDir()+"/gitconfig")
	setGitConfig(t, "https.proxy", "http://127.0.0.1:7890")

	hint, err := brewProxyHint()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Homebrew/curl", "siti proxy on", "siti proxy git off"} {
		if !strings.Contains(hint, want) {
			t.Errorf("hint %q does not contain %q", hint, want)
		}
	}
}

func TestBrewProxyHintHiddenWhenTerminalProxyIsEnabled(t *testing.T) {
	clearProxyEnvironment(t)
	t.Setenv("GIT_CONFIG_GLOBAL", t.TempDir()+"/gitconfig")
	setGitConfig(t, "https.proxy", "http://127.0.0.1:7890")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:7890")

	hint, err := brewProxyHint()
	if err != nil {
		t.Fatal(err)
	}
	if hint != "" {
		t.Fatalf("unexpected hint: %q", hint)
	}
}

func clearProxyEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{"http_proxy", "HTTP_PROXY", "https_proxy", "HTTPS_PROXY", "all_proxy", "ALL_PROXY"} {
		t.Setenv(key, "")
	}
}

func setGitConfig(t *testing.T, key, value string) {
	t.Helper()
	if out, err := exec.Command("git", "config", "--global", key, value).CombinedOutput(); err != nil {
		t.Fatalf("git config %s: %v: %s", key, err, out)
	}
}
