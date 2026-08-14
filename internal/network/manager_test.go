package network

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type fakeRunner struct {
	calls      []string
	outputs    map[string]string
	runErrors  map[string]error
	outputErrs map[string]error
	onRun      func(string)
}

func (f *fakeRunner) Run(name string, args ...string) error {
	key := commandKey(name, args...)
	f.calls = append(f.calls, key)
	if f.onRun != nil {
		f.onRun(key)
	}
	return f.runErrors[key]
}

func (f *fakeRunner) Output(name string, args ...string) (string, error) {
	key := commandKey(name, args...)
	f.calls = append(f.calls, key)
	return f.outputs[key], f.outputErrs[key]
}

func commandKey(name string, args ...string) string {
	return strings.Join(append([]string{name}, args...), " ")
}

func newTestManager(t *testing.T) (*Manager, *fakeRunner) {
	t.Helper()
	dir := t.TempDir()
	networkSetup := filepath.Join(dir, "networksetup")
	if err := os.WriteFile(networkSetup, []byte("test"), 0o700); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{
		outputs: map[string]string{
			commandKey(networkSetup, "-listallhardwareports"): `Hardware Port: Ethernet
Device: en0

Hardware Port: Wi-Fi
Device: en7
`,
			commandKey(networkSetup, "-listnetworkserviceorder"): `(1) Ethernet
(Hardware Port: Ethernet, Device: en0)
(2) Office Wireless
(Hardware Port: Wi-Fi, Device: en7)
`,
			commandKey(networkSetup, "-getinfo", "Office Wireless"): `Manual Configuration
IP address: 172.16.40.100
Subnet mask: 255.255.248.0
Router: 172.16.40.2
`,
			commandKey(networkSetup, "-getdnsservers", "Office Wireless"): "172.16.40.2\n",
		},
		runErrors:  map[string]error{},
		outputErrs: map[string]error{},
	}
	return &Manager{
		ConfigDir:    dir,
		goos:         "darwin",
		networkSetup: networkSetup,
		runner:       runner,
	}, runner
}

func TestApplyUsesBuiltinProfileAndRenamedWiFiService(t *testing.T) {
	manager, runner := newTestManager(t)
	result, err := manager.Apply("blacklake-proxy")
	if err != nil {
		t.Fatal(err)
	}
	if result.State.Device != "en7" || result.State.Service != "Office Wireless" {
		t.Fatalf("state = %+v", result.State)
	}

	wantRuns := []string{
		"sudo -v",
		"sudo " + manager.networkSetup + " -setmanual Office Wireless 172.16.40.100 255.255.248.0 172.16.40.2",
		"sudo " + manager.networkSetup + " -setdnsservers Office Wireless 172.16.40.2",
	}
	for _, want := range wantRuns {
		if !contains(runner.calls, want) {
			t.Fatalf("missing call %q in %v", want, runner.calls)
		}
	}
	active, ok, err := ReadActive(manager.ConfigDir)
	if err != nil || !ok || active.Profile != "blacklake-proxy" {
		t.Fatalf("active = %+v, ok = %v, err = %v", active, ok, err)
	}
}

func TestApplyBuiltinUsesCurrentWiFiAddress(t *testing.T) {
	manager, runner := newTestManager(t)
	infoKey := commandKey(manager.networkSetup, "-getinfo", "Office Wireless")
	switchCall := "sudo " + manager.networkSetup + " -setairportnetwork en7 blacklake"
	runner.outputs[infoKey] = `DHCP Configuration
IP address: 172.16.40.141
Subnet mask: 255.255.255.0
Router: 172.16.40.1
`
	manualCall := "sudo " + manager.networkSetup + " -setmanual Office Wireless 172.16.40.141 255.255.248.0 172.16.40.2"
	runner.onRun = func(call string) {
		if call == manualCall {
			runner.outputs[infoKey] = `Manual Configuration
IP address: 172.16.40.141
Subnet mask: 255.255.248.0
Router: 172.16.40.2
`
		}
	}

	result, err := manager.Apply("blacklake-proxy")
	if err != nil {
		t.Fatal(err)
	}
	if result.State.IPv4.Address != "172.16.40.141" {
		t.Fatalf("address = %q", result.State.IPv4.Address)
	}
	if !contains(runner.calls, manualCall) {
		t.Fatalf("missing call %q in %v", manualCall, runner.calls)
	}
	if contains(runner.calls, switchCall) {
		t.Fatalf("unexpected Wi-Fi switch: %v", runner.calls)
	}
}

func TestApplyBuiltinRejectsMissingCurrentWiFiAddress(t *testing.T) {
	manager, runner := newTestManager(t)
	runner.outputs[commandKey(manager.networkSetup, "-getinfo", "Office Wireless")] = "DHCP Configuration\n"

	_, err := manager.Apply("blacklake-proxy")
	if err == nil || !strings.Contains(err.Error(), "没有可用的 IPv4") || !strings.Contains(err.Error(), "手动连接") {
		t.Fatalf("err = %v", err)
	}
	for _, call := range runner.calls {
		if strings.Contains(call, " -setmanual ") || strings.Contains(call, " -setdnsservers ") {
			t.Fatalf("unexpected network configuration call: %s", call)
		}
	}
}

func TestApplyBuiltinRejectsUnexpectedSubnet(t *testing.T) {
	manager, runner := newTestManager(t)
	runner.outputs[commandKey(manager.networkSetup, "-getinfo", "Office Wireless")] = `DHCP Configuration
IP address: 192.168.101.143
Subnet mask: 255.255.255.0
Router: 192.168.101.1
`

	_, err := manager.Apply("blacklake-proxy")
	if err == nil || !strings.Contains(err.Error(), "不在 blacklake 的目标网段") || !strings.Contains(err.Error(), "手动连接") {
		t.Fatalf("err = %v", err)
	}
	for _, call := range runner.calls {
		if strings.HasPrefix(call, "sudo ") || strings.Contains(call, " -setairportnetwork ") {
			t.Fatalf("unexpected privileged or Wi-Fi switch call: %s", call)
		}
	}
}

func TestApplyIsIdempotentWhenProfileMatches(t *testing.T) {
	manager, runner := newTestManager(t)
	if _, err := manager.Apply("blacklake-proxy"); err != nil {
		t.Fatal(err)
	}
	runner.calls = nil

	result, err := manager.Apply("blacklake-proxy")
	if err != nil {
		t.Fatal(err)
	}
	if !result.AlreadyApplied {
		t.Fatal("expected already applied")
	}
	for _, call := range runner.calls {
		if strings.HasPrefix(call, "sudo ") {
			t.Fatalf("unexpected privileged call: %s", call)
		}
	}
}

func TestResetRestoresDHCPAndAutomaticDNS(t *testing.T) {
	manager, runner := newTestManager(t)
	if _, err := manager.Apply("blacklake-proxy"); err != nil {
		t.Fatal(err)
	}
	runner.calls = nil
	runner.outputs[commandKey(manager.networkSetup, "-getinfo", "Office Wireless")] = `DHCP Configuration
IP address: 172.16.40.101
Subnet mask: 255.255.255.0
Router: 172.16.40.2
`
	runner.outputs[commandKey(manager.networkSetup, "-getdnsservers", "Office Wireless")] = "There aren't any DNS Servers set on Office Wireless.\n"

	result, err := manager.Reset()
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed {
		t.Fatal("expected reset to change state")
	}
	if result.Service != "Office Wireless" {
		t.Fatalf("service = %q", result.Service)
	}
	wantLive := LiveStatus{
		Mode:       "DHCP Configuration",
		Address:    "172.16.40.101",
		SubnetMask: "255.255.255.0",
		Gateway:    "172.16.40.2",
	}
	if !reflect.DeepEqual(result.Live, wantLive) {
		t.Fatalf("live = %+v, want %+v", result.Live, wantLive)
	}
	want := []string{
		"sudo -v",
		"sudo " + manager.networkSetup + " -setdhcp Office Wireless Empty",
		"sudo " + manager.networkSetup + " -setdnsservers Office Wireless Empty",
	}
	for _, call := range want {
		if !contains(runner.calls, call) {
			t.Fatalf("missing call %q in %v", call, runner.calls)
		}
	}
	if _, active, err := ReadActive(manager.ConfigDir); err != nil || active {
		t.Fatalf("active = %v, err = %v", active, err)
	}
}

func TestApplyRollsBackWhenDNSFails(t *testing.T) {
	manager, runner := newTestManager(t)
	writeProfile(t, manager.ConfigDir, "blacklake-proxy", validProfileYAML)
	dnsCall := "sudo " + manager.networkSetup + " -setdnsservers Office Wireless 172.16.40.2"
	runner.runErrors[dnsCall] = errors.New("dns failed")
	runner.outputs[commandKey(manager.networkSetup, "-getinfo", "Office Wireless")] = "DHCP Configuration\n"
	runner.outputs[commandKey(manager.networkSetup, "-getdnsservers", "Office Wireless")] = "There aren't any DNS Servers set on Office Wireless.\n"

	_, err := manager.Apply("blacklake-proxy")
	if err == nil || !strings.Contains(err.Error(), "已恢复 DHCP") {
		t.Fatalf("err = %v", err)
	}
	if _, active, readErr := ReadActive(manager.ConfigDir); readErr != nil || active {
		t.Fatalf("active = %v, err = %v", active, readErr)
	}
}

func TestResetWithoutActiveProfileIsNoOp(t *testing.T) {
	manager, runner := newTestManager(t)
	result, err := manager.Reset()
	if err != nil {
		t.Fatal(err)
	}
	if result.Changed || len(runner.calls) != 0 {
		t.Fatalf("result = %+v, calls = %v", result, runner.calls)
	}
}

func TestNetworkServiceDisabled(t *testing.T) {
	manager, runner := newTestManager(t)
	runner.outputs[commandKey(manager.networkSetup, "-listnetworkserviceorder")] = `(*) Office Wireless
(Hardware Port: Wi-Fi, Device: en7)
`
	_, err := manager.Apply("blacklake-proxy")
	if err == nil || !strings.Contains(err.Error(), "已被禁用") {
		t.Fatalf("err = %v", err)
	}
}

func TestParseNetworkInfo(t *testing.T) {
	got := parseNetworkInfo("Manual Configuration\nIP address: 10.0.0.2\nSubnet mask: 255.255.255.0\nRouter: 10.0.0.1\n")
	want := LiveStatus{Mode: "Manual Configuration", Address: "10.0.0.2", SubnetMask: "255.255.255.0", Gateway: "10.0.0.1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestAddressSharesSubnet(t *testing.T) {
	if !addressSharesSubnet("172.16.40.229", "172.16.40.2", "255.255.248.0") {
		t.Fatal("expected addresses to share /21 subnet")
	}
	if addressSharesSubnet("192.168.101.143", "172.16.40.2", "255.255.248.0") {
		t.Fatal("unexpected subnet match")
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
