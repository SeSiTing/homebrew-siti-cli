package network

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const defaultNetworkSetup = "/usr/sbin/networksetup"

const (
	defaultRouteCommand = "/sbin/route"
	defaultDigCommand   = "/usr/bin/dig"
	defaultCurlCommand  = "/usr/bin/curl"
	connectivityHost    = "github.com"
)

type commandRunner interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) (string, error)
}

type execCommandRunner struct{}

func (execCommandRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (execCommandRunner) Output(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(out))
		if message != "" {
			return "", fmt.Errorf("%s: %w", message, err)
		}
		return "", err
	}
	return string(out), nil
}

type Manager struct {
	ConfigDir    string
	goos         string
	networkSetup string
	routeCommand string
	digCommand   string
	curlCommand  string
	sleep        func(time.Duration)
	runner       commandRunner
}

type ApplyResult struct {
	State          ActiveState
	Live           LiveStatus
	Checks         ApplyChecks
	AlreadyApplied bool
}

type ApplyChecks struct {
	Gateway  string
	DNS      string
	Internet string
}

type ResetResult struct {
	Service string
	Device  string
	Live    LiveStatus
}

type LiveStatus struct {
	Mode       string
	Address    string
	SubnetMask string
	Gateway    string
	DNS        []string
}

type StatusResult struct {
	Active  bool
	State   ActiveState
	Service string
	Live    LiveStatus
}

func NewManager() (*Manager, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return nil, err
	}
	return &Manager{
		ConfigDir:    dir,
		goos:         runtime.GOOS,
		networkSetup: defaultNetworkSetup,
		routeCommand: defaultRouteCommand,
		digCommand:   defaultDigCommand,
		curlCommand:  defaultCurlCommand,
		sleep:        time.Sleep,
		runner:       execCommandRunner{},
	}, nil
}

func (m *Manager) Apply(name string) (ApplyResult, error) {
	profile, err := ReadProfile(m.ConfigDir, name)
	if err != nil {
		return ApplyResult{}, err
	}
	if err := m.checkSupported(); err != nil {
		return ApplyResult{}, err
	}
	device, service, err := m.resolveWiFi()
	if err != nil {
		return ApplyResult{}, err
	}
	active, hasActive, err := ReadActive(m.ConfigDir)
	if err != nil {
		return ApplyResult{}, err
	}
	if profile.CurrentAddress || profile.CurrentSubnetMask {
		allowManual := hasActive && active.Profile == name
		live, err := m.prepareCurrentNetwork(profile, service, allowManual)
		if err != nil {
			return ApplyResult{}, err
		}
		if profile.CurrentAddress {
			profile.IPv4.Address = live.Address
			profile.CurrentAddress = false
		}
		if profile.CurrentSubnetMask {
			profile.IPv4.SubnetMask = live.SubnetMask
			profile.CurrentSubnetMask = false
		}
	}

	state := ActiveState{
		Profile:   name,
		Interface: profile.Interface,
		SSID:      profile.SSID,
		Device:    device,
		Service:   service,
		IPv4:      profile.IPv4,
		DNS:       append([]string(nil), profile.DNS...),
	}
	if hasActive && active.Profile == name {
		live, err := m.readLive(service)
		if err == nil && matchesProfile(live, profile) {
			checks, verifyErr := m.verifyConnectivity(profile, device)
			if verifyErr == nil {
				return ApplyResult{State: state, Live: live, Checks: checks, AlreadyApplied: true}, nil
			}
		}
	}

	if err := m.runner.Run("sudo", "-v"); err != nil {
		return ApplyResult{}, fmt.Errorf("获取管理员权限: %w", err)
	}
	if hasActive && active.Profile != name {
		if _, _, err := m.resetState(active); err != nil {
			return ApplyResult{}, fmt.Errorf("重置当前 profile %q: %w", active.Profile, err)
		}
		if err := RemoveActive(m.ConfigDir); err != nil {
			return ApplyResult{}, err
		}
	}

	if err := WriteActive(m.ConfigDir, state); err != nil {
		return ApplyResult{}, err
	}
	if err := m.runNetworkSetup("-setmanual", service, profile.IPv4.Address, profile.IPv4.SubnetMask, profile.IPv4.Gateway); err != nil {
		_, rollbackErr := m.resetService(service)
		if rollbackErr == nil {
			_ = RemoveActive(m.ConfigDir)
			return ApplyResult{}, fmt.Errorf("设置固定 IPv4: %w（已恢复 DHCP 和自动 DNS）", err)
		}
		return ApplyResult{}, fmt.Errorf("设置固定 IPv4: %w；自动回滚失败: %v", err, rollbackErr)
	}
	dnsArgs := append([]string{"-setdnsservers", service}, profile.DNS...)
	if err := m.runNetworkSetup(dnsArgs...); err != nil {
		_, rollbackErr := m.resetService(service)
		if rollbackErr == nil {
			_ = RemoveActive(m.ConfigDir)
			return ApplyResult{}, fmt.Errorf("设置 DNS: %w（已恢复 DHCP 和自动 DNS）", err)
		}
		return ApplyResult{}, fmt.Errorf("设置 DNS: %w；自动回滚失败: %v", err, rollbackErr)
	}
	live, err := m.waitForProfile(service, profile)
	if err != nil {
		_, rollbackErr := m.resetService(service)
		if rollbackErr == nil {
			_ = RemoveActive(m.ConfigDir)
			return ApplyResult{}, fmt.Errorf("%w（已恢复 DHCP 和自动 DNS）", err)
		}
		return ApplyResult{}, fmt.Errorf("%w；自动回滚失败: %v", err, rollbackErr)
	}
	checks, err := m.verifyConnectivity(profile, device)
	if err != nil {
		_, rollbackErr := m.resetService(service)
		if rollbackErr == nil {
			_ = RemoveActive(m.ConfigDir)
			return ApplyResult{}, fmt.Errorf("网络可用性校验失败: %w（已恢复 DHCP 和自动 DNS）", err)
		}
		return ApplyResult{}, fmt.Errorf("网络可用性校验失败: %w；自动回滚失败: %v", err, rollbackErr)
	}
	return ApplyResult{State: state, Live: live, Checks: checks}, nil
}

func (m *Manager) waitForProfile(service string, profile Profile) (LiveStatus, error) {
	const attempts = 6
	var last LiveStatus
	var lastErr error
	stableReads := 0
	for attempt := 0; attempt < attempts; attempt++ {
		live, err := m.readLive(service)
		if err == nil && matchesProfile(live, profile) {
			last = live
			stableReads++
			if stableReads == 2 {
				return live, nil
			}
		} else {
			stableReads = 0
			last = live
			lastErr = err
		}
		if attempt < attempts-1 {
			m.sleep(500 * time.Millisecond)
		}
	}
	if lastErr != nil {
		return LiveStatus{}, fmt.Errorf("读回网络配置: %w", lastErr)
	}
	return LiveStatus{}, fmt.Errorf("网络配置未稳定生效（当前 IPv4: %s/%s, gateway: %s, DNS: %s）",
		last.Address, last.SubnetMask, last.Gateway, displayDNS(last.DNS))
}

func (m *Manager) verifyConnectivity(profile Profile, device string) (ApplyChecks, error) {
	gateway, err := m.waitForDefaultRoute(profile.IPv4.Gateway, device)
	if err != nil {
		return ApplyChecks{}, err
	}

	dnsServer := profile.DNS[0]
	digOut, err := m.runner.Output(m.digCommand, connectivityDigArgs(dnsServer)...)
	if err != nil {
		return ApplyChecks{}, fmt.Errorf("DNS %s 查询失败: %w", dnsServer, err)
	}
	resolvedIP := firstIPv4(digOut)
	if resolvedIP == "" {
		return ApplyChecks{}, fmt.Errorf("DNS %s 无法解析 %s", dnsServer, connectivityHost)
	}

	_, err = m.runner.Output(m.curlCommand, connectivityCurlArgs(resolvedIP)...)
	if err != nil {
		return ApplyChecks{}, fmt.Errorf("无法通过软路由访问 %s: %w", connectivityHost, err)
	}

	return ApplyChecks{Gateway: gateway, DNS: dnsServer, Internet: connectivityHost}, nil
}

func (m *Manager) waitForDefaultRoute(expectedGateway, expectedDevice string) (string, error) {
	const attempts = 16
	var gateway string
	var device string
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		routeOut, err := m.runner.Output(m.routeCommand, "-n", "get", "default")
		if err == nil {
			gateway, device = parseDefaultRoute(routeOut)
			if gateway == expectedGateway && device == expectedDevice {
				return gateway, nil
			}
		}
		lastErr = err
		if attempt < attempts-1 {
			m.sleep(500 * time.Millisecond)
		}
	}
	if gateway == "" && device == "" && lastErr != nil {
		return "", fmt.Errorf("读取默认路由: %w", lastErr)
	}
	return "", fmt.Errorf("默认路由不正确（当前: %s via %s，预期: %s via %s）",
		displayValue(gateway), displayValue(device), expectedGateway, expectedDevice)
}

func connectivityDigArgs(dnsServer string) []string {
	return []string{"+time=2", "+tries=1", "+short", "@" + dnsServer, connectivityHost, "A"}
}

func connectivityCurlArgs(resolvedIP string) []string {
	resolve := fmt.Sprintf("%s:443:%s", connectivityHost, resolvedIP)
	return []string{
		"--noproxy", "*",
		"--connect-timeout", "3",
		"--max-time", "8",
		"--silent",
		"--show-error",
		"--output", "/dev/null",
		"--resolve", resolve,
		"https://" + connectivityHost,
	}
}

func (m *Manager) prepareCurrentNetwork(profile Profile, service string, allowManual bool) (LiveStatus, error) {
	live, err := m.readLive(service)
	if err != nil {
		return LiveStatus{}, fmt.Errorf("读取当前 Wi-Fi 配置: %w", err)
	}
	if live.Mode != "DHCP Configuration" && !allowManual {
		return LiveStatus{}, fmt.Errorf("当前 Wi-Fi 不是 DHCP 模式（%s）；请先运行 siti net reset", displayValue(live.Mode))
	}
	if profile.CurrentAddress && !isIPv4(live.Address) {
		if profile.SSID != "" {
			return LiveStatus{}, fmt.Errorf("当前 Wi-Fi 没有可用的 IPv4 地址；请先手动连接 %q", profile.SSID)
		}
		return LiveStatus{}, fmt.Errorf("当前 Wi-Fi 没有可用的 IPv4 地址")
	}
	if profile.CurrentSubnetMask && !isSubnetMask(live.SubnetMask) {
		return LiveStatus{}, fmt.Errorf("当前 Wi-Fi 没有有效的子网掩码: %s", displayValue(live.SubnetMask))
	}
	address := profile.IPv4.Address
	if profile.CurrentAddress {
		address = live.Address
	}
	subnetMask := profile.IPv4.SubnetMask
	if profile.CurrentSubnetMask {
		subnetMask = live.SubnetMask
	}
	if !addressSharesSubnet(address, profile.IPv4.Gateway, subnetMask) {
		if profile.SSID != "" {
			return LiveStatus{}, fmt.Errorf("当前 Wi-Fi 地址 %s 无法访问 %s 网关（子网掩码 %s）；请先手动连接 %q", address, profile.IPv4.Gateway, subnetMask, profile.SSID)
		}
		return LiveStatus{}, fmt.Errorf("当前 Wi-Fi 地址 %s 无法访问 %s 网关（子网掩码 %s）", address, profile.IPv4.Gateway, subnetMask)
	}
	return live, nil
}

func addressSharesSubnet(address, gateway, subnetMask string) bool {
	addressIP := net.ParseIP(address).To4()
	gatewayIP := net.ParseIP(gateway).To4()
	maskIP := net.ParseIP(subnetMask).To4()
	if addressIP == nil || gatewayIP == nil || maskIP == nil {
		return false
	}
	mask := net.IPMask(maskIP)
	return addressIP.Mask(mask).Equal(gatewayIP.Mask(mask))
}

func (m *Manager) Reset() (ResetResult, error) {
	if err := m.checkSupported(); err != nil {
		return ResetResult{}, err
	}
	device, service, err := m.resolveWiFi()
	if err != nil {
		return ResetResult{}, err
	}
	if err := m.runner.Run("sudo", "-v"); err != nil {
		return ResetResult{}, fmt.Errorf("获取管理员权限: %w", err)
	}
	live, err := m.resetService(service)
	if err != nil {
		return ResetResult{}, err
	}
	if err := RemoveActive(m.ConfigDir); err != nil {
		return ResetResult{}, err
	}
	return ResetResult{Service: service, Device: device, Live: live}, nil
}

func (m *Manager) Status() (StatusResult, error) {
	state, active, err := ReadActive(m.ConfigDir)
	if err != nil || !active {
		return StatusResult{Active: active}, err
	}
	if err := m.checkSupported(); err != nil {
		return StatusResult{}, err
	}
	service, err := m.serviceForDevice(state.Device)
	if err != nil {
		return StatusResult{}, err
	}
	live, err := m.readLive(service)
	if err != nil {
		return StatusResult{}, err
	}
	return StatusResult{Active: true, State: state, Service: service, Live: live}, nil
}

func (m *Manager) checkSupported() error {
	if m.goos != "darwin" {
		return fmt.Errorf("network profile 目前仅支持 macOS")
	}
	if _, err := os.Stat(m.networkSetup); err != nil {
		return fmt.Errorf("未找到 networksetup: %s", m.networkSetup)
	}
	return nil
}

func (m *Manager) resolveWiFi() (string, string, error) {
	out, err := m.runner.Output(m.networkSetup, "-listallhardwareports")
	if err != nil {
		return "", "", fmt.Errorf("读取网络硬件端口: %w", err)
	}
	device := wifiDevice(out)
	if device == "" {
		return "", "", fmt.Errorf("没有找到 Wi-Fi 网络接口")
	}
	service, err := m.serviceForDevice(device)
	if err != nil {
		return "", "", err
	}
	return device, service, nil
}

func (m *Manager) serviceForDevice(device string) (string, error) {
	out, err := m.runner.Output(m.networkSetup, "-listnetworkserviceorder")
	if err != nil {
		return "", fmt.Errorf("读取 network service: %w", err)
	}
	service, disabled := networkService(out, device)
	if service == "" {
		return "", fmt.Errorf("没有找到接口 %s 对应的 network service", device)
	}
	if disabled {
		return "", fmt.Errorf("network service %q 已被禁用", service)
	}
	return service, nil
}

func (m *Manager) resetState(state ActiveState) (string, LiveStatus, error) {
	service, err := m.serviceForDevice(state.Device)
	if err != nil {
		return "", LiveStatus{}, err
	}
	live, err := m.resetService(service)
	return service, live, err
}

func (m *Manager) resetService(service string) (LiveStatus, error) {
	if err := m.runNetworkSetup("-setdhcp", service); err != nil {
		return LiveStatus{}, fmt.Errorf("恢复 DHCP: %w", err)
	}
	if err := m.runNetworkSetup("-setdnsservers", service, "Empty"); err != nil {
		return LiveStatus{}, fmt.Errorf("恢复自动 DNS: %w", err)
	}
	live, err := m.readLive(service)
	if err != nil {
		return LiveStatus{}, fmt.Errorf("读回网络配置: %w", err)
	}
	if live.Mode != "DHCP Configuration" || len(live.DNS) != 0 {
		return LiveStatus{}, fmt.Errorf("恢复后的网络配置不是 DHCP 和自动 DNS")
	}
	return live, nil
}

func (m *Manager) runNetworkSetup(args ...string) error {
	if err := m.runner.Run("sudo", append([]string{m.networkSetup}, args...)...); err != nil {
		return err
	}
	return nil
}

func (m *Manager) readLive(service string) (LiveStatus, error) {
	info, err := m.runner.Output(m.networkSetup, "-getinfo", service)
	if err != nil {
		return LiveStatus{}, err
	}
	dns, err := m.runner.Output(m.networkSetup, "-getdnsservers", service)
	if err != nil {
		return LiveStatus{}, err
	}
	live := parseNetworkInfo(info)
	live.DNS = parseDNS(dns)
	return live, nil
}

func wifiDevice(output string) string {
	isWiFi := false
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "Hardware Port: ") {
			port := strings.TrimPrefix(line, "Hardware Port: ")
			isWiFi = port == "Wi-Fi" || port == "AirPort"
			continue
		}
		if isWiFi && strings.HasPrefix(line, "Device: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Device: "))
		}
	}
	return ""
}

func networkService(output, device string) (string, bool) {
	service := ""
	disabled := false
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if name, isDisabled, ok := parseServiceLine(line); ok {
			service, disabled = name, isDisabled
			continue
		}
		if service != "" && strings.Contains(line, "Device: "+device+")") {
			return service, disabled
		}
	}
	return "", false
}

func parseServiceLine(line string) (string, bool, bool) {
	if !strings.HasPrefix(line, "(") {
		return "", false, false
	}
	end := strings.Index(line, ") ")
	if end < 2 {
		return "", false, false
	}
	marker := line[1:end]
	for _, char := range marker {
		if (char < '0' || char > '9') && char != '*' {
			return "", false, false
		}
	}
	name := strings.TrimSpace(line[end+2:])
	return name, marker == "*", name != ""
}

func parseNetworkInfo(output string) LiveStatus {
	status := LiveStatus{}
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if status.Mode == "" {
			status.Mode = line
			continue
		}
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			continue
		}
		switch key {
		case "IP address":
			status.Address = value
		case "Subnet mask":
			status.SubnetMask = value
		case "Router":
			status.Gateway = value
		}
	}
	return status
}

func parseDNS(output string) []string {
	var servers []string
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.Contains(strings.ToLower(line), "aren't any dns servers") {
			continue
		}
		servers = append(servers, line)
	}
	return servers
}

func parseDefaultRoute(output string) (string, string) {
	var gateway string
	var device string
	for _, raw := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(raw), ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "gateway":
			gateway = strings.TrimSpace(value)
		case "interface":
			device = strings.TrimSpace(value)
		}
	}
	return gateway, device
}

func firstIPv4(output string) string {
	for _, field := range strings.Fields(output) {
		if ip := net.ParseIP(strings.TrimSpace(field)); ip != nil && ip.To4() != nil {
			return ip.String()
		}
	}
	return ""
}

func displayDNS(servers []string) string {
	if len(servers) == 0 {
		return "none"
	}
	return strings.Join(servers, ", ")
}

func displayValue(value string) string {
	if value == "" {
		return "none"
	}
	return value
}

func matchesProfile(live LiveStatus, profile Profile) bool {
	return live.Mode == "Manual Configuration" &&
		live.Address == profile.IPv4.Address &&
		live.SubnetMask == profile.IPv4.SubnetMask &&
		live.Gateway == profile.IPv4.Gateway &&
		equalStrings(live.DNS, profile.DNS)
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
