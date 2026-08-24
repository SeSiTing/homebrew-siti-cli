package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type clashProcessStatus struct {
	appRunning    bool
	helperRunning bool
	coreRunning   bool
	configPath    string
}

type clashTunConfig struct {
	enabled bool
	device  string
}

type clashRuntimeStatus struct {
	processKnown          bool
	appRunning            bool
	helperRunning         bool
	coreRunning           bool
	portOpen              bool
	systemProxyKnown      bool
	systemProxyConfigured bool
	tunConfigKnown        bool
	tunConfigured         bool
	tunDevice             string
	tunInterfaceReady     bool
	tunActive             bool
}

var clashConfigFlagPattern = regexp.MustCompile(`(?:^|\s)-f(?:=|\s+)(.*?)(?:\s+-[A-Za-z][A-Za-z0-9-]*(?:=|\s|$)|$)`)

func detectClashRuntime(systemProxies map[string]string, systemProxyKnown bool) clashRuntimeStatus {
	status := clashRuntimeStatus{systemProxyKnown: systemProxyKnown}

	if out, err := exec.Command("ps", "-axo", "command=").CombinedOutput(); err == nil {
		processes := parseClashProcesses(string(out))
		status.processKnown = true
		status.appRunning = processes.appRunning
		status.helperRunning = processes.helperRunning
		status.coreRunning = processes.coreRunning

		if config, known := loadClashTunConfig(clashConfigCandidates(processes.configPath)); known {
			status.tunConfigKnown = true
			status.tunConfigured = config.enabled
			status.tunDevice = config.device
		}
	}

	status.portOpen = tcpEndpointOpen(net.JoinHostPort(proxyHost, proxyPort))
	status.systemProxyConfigured = clashSystemProxyConfigured(systemProxies)
	status.tunInterfaceReady, status.tunActive = resolveClashTun(status)
	return status
}

func parseClashProcesses(output string) clashProcessStatus {
	var status clashProcessStatus
	for _, command := range strings.Split(output, "\n") {
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}
		lower := strings.ToLower(command)

		if strings.Contains(lower, "party.mihomo.helper") ||
			strings.Contains(lower, "clash-verge-service") ||
			strings.Contains(lower, "clash-core-service") {
			status.helperRunning = true
			continue
		}

		if strings.Contains(lower, "/applications/clash verge.app/contents/macos/clash-verge") {
			status.appRunning = true
			continue
		}

		if !isClashCoreCommand(command) {
			continue
		}
		status.coreRunning = true
		if status.configPath == "" {
			status.configPath = clashConfigPathFromCommand(command)
		}
	}
	return status
}

func isClashCoreCommand(command string) bool {
	lower := strings.ToLower(command)
	if strings.Contains(lower, "verge-mihomo") {
		return true
	}
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return false
	}
	switch strings.ToLower(filepath.Base(fields[0])) {
	case "mihomo", "clash", "clash-meta":
		return true
	default:
		return false
	}
}

func clashConfigPathFromCommand(command string) string {
	match := clashConfigFlagPattern.FindStringSubmatch(command)
	if len(match) != 2 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(match[1]), `"'`)
}

func clashConfigCandidates(processPath string) []string {
	var candidates []string
	if processPath != "" {
		candidates = append(candidates, processPath)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, "Library", "Application Support", "io.github.clash-verge-rev.clash-verge-rev", "clash-verge.yaml"),
			filepath.Join(home, ".config", "clash", "config.yaml"),
		)
	}
	return uniqueStrings(candidates)
}

func loadClashTunConfig(paths []string) (clashTunConfig, bool) {
	type tunBlock struct {
		Enable bool   `yaml:"enable"`
		Device string `yaml:"device"`
	}
	type clashConfig struct {
		Tun *tunBlock `yaml:"tun"`
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var config clashConfig
		if err := yaml.Unmarshal(data, &config); err != nil || config.Tun == nil {
			continue
		}
		return clashTunConfig{
			enabled: config.Tun.Enable,
			device:  strings.TrimSpace(config.Tun.Device),
		}, true
	}
	return clashTunConfig{}, false
}

func clashSystemProxyConfigured(proxies map[string]string) bool {
	for _, name := range []string{"HTTP", "HTTPS", "SOCKS"} {
		endpoint, ok := localProxyEndpoint(proxies[name])
		if !ok {
			continue
		}
		host, port, err := net.SplitHostPort(endpoint)
		if err != nil || port != proxyPort {
			continue
		}
		if strings.EqualFold(host, "localhost") {
			return true
		}
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			return true
		}
	}
	return false
}

func resolveClashTun(status clashRuntimeStatus) (interfaceReady, active bool) {
	if !status.coreRunning || !status.tunConfigKnown || !status.tunConfigured {
		return false, false
	}
	if !strings.HasPrefix(strings.ToLower(status.tunDevice), "utun") {
		return false, true
	}
	iface, err := net.InterfaceByName(status.tunDevice)
	if err != nil || iface.Flags&net.FlagUp == 0 {
		return false, false
	}
	return true, true
}

func tcpEndpointOpen(endpoint string) bool {
	connection, err := net.DialTimeout("tcp", endpoint, 300*time.Millisecond)
	if connection != nil {
		_ = connection.Close()
	}
	return err == nil
}

func clashStatusLines(status clashRuntimeStatus) []string {
	lines := []string{"本机 Clash:"}
	if !status.processKnown {
		lines = append(lines, "  运行状态:   未知（进程检查失败）")
	} else {
		lines = append(lines,
			fmt.Sprintf("  主程序:     %s", runningText(status.appRunning)),
			fmt.Sprintf("  后台服务:   %s", helperRunningText(status.helperRunning)),
			fmt.Sprintf("  代理核心:   %s", runningText(status.coreRunning)),
		)
	}

	lines = append(lines,
		fmt.Sprintf("  系统代理:   %s", clashSystemProxyText(status)),
		fmt.Sprintf("  虚拟网卡:   %s", clashTunText(status)),
		fmt.Sprintf("  本地端口:   %s", clashPortText(status.portOpen)),
		fmt.Sprintf("  当前接管:   %s", clashModeText(status)),
	)
	return lines
}

func runningText(running bool) string {
	if running {
		return "运行中"
	}
	return "未运行"
}

func helperRunningText(running bool) string {
	if running {
		return "运行中（不代表代理已开启）"
	}
	return "未运行"
}

func clashSystemProxyText(status clashRuntimeStatus) string {
	if !status.systemProxyKnown {
		return "未知"
	}
	if !status.systemProxyConfigured {
		return "关闭"
	}
	if !status.portOpen {
		return fmt.Sprintf("已配置（%s:%s 未监听）", proxyHost, proxyPort)
	}
	return fmt.Sprintf("开启（%s:%s）", proxyHost, proxyPort)
}

func clashTunText(status clashRuntimeStatus) string {
	if !status.tunConfigKnown {
		if status.coreRunning {
			return "未知（无法读取运行配置）"
		}
		return "关闭"
	}
	if !status.tunConfigured {
		return "关闭"
	}
	if !status.coreRunning {
		return "关闭（配置已开启，代理核心未运行）"
	}
	if strings.HasPrefix(strings.ToLower(status.tunDevice), "utun") && !status.tunInterfaceReady {
		return fmt.Sprintf("异常（%s 未就绪）", status.tunDevice)
	}
	if status.tunDevice != "" {
		return fmt.Sprintf("开启（%s）", status.tunDevice)
	}
	return "开启"
}

func clashPortText(open bool) string {
	if open {
		return proxyPort + " open"
	}
	return proxyPort + " closed"
}

func clashModeText(status clashRuntimeStatus) string {
	var modes []string
	if status.systemProxyConfigured && status.portOpen {
		modes = append(modes, "系统代理")
	}
	if status.tunActive {
		modes = append(modes, "虚拟网卡")
	}
	if len(modes) == 0 {
		return "未接管"
	}
	return strings.Join(modes, " + ")
}

func (status clashRuntimeStatus) active() bool {
	return status.systemProxyConfigured && status.portOpen || status.tunActive
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
