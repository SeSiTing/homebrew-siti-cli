package network

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const validProfileYAML = `version: 1
interface: wifi
ipv4:
  address: 172.16.40.100
  subnet_mask: 255.255.255.0
  gateway: 172.16.40.2
dns:
  - 172.16.40.2
`

const currentAddressProfileYAML = `version: 1
interface: wifi
ssid: blacklake
ipv4:
  address: current
  subnet_mask: 255.255.248.0
  gateway: 172.16.40.2
dns:
  - 172.16.40.2
`

func writeProfile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestReadProfile(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "blacklake-proxy", validProfileYAML)

	profile, err := ReadProfile(dir, "blacklake-proxy")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Interface != "wifi" || profile.IPv4.Address != "172.16.40.100" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if !reflect.DeepEqual(profile.DNS, []string{"172.16.40.2"}) {
		t.Fatalf("DNS = %v", profile.DNS)
	}
}

func TestReadBuiltinBlacklakeProxy(t *testing.T) {
	profile, err := ReadProfile(t.TempDir(), "blacklake-proxy")
	if err != nil {
		t.Fatal(err)
	}
	if !profile.CurrentAddress || profile.IPv4.Address != "" {
		t.Fatalf("address mode = %+v", profile)
	}
	if profile.SSID != "blacklake" {
		t.Fatalf("SSID = %q", profile.SSID)
	}
	if profile.IPv4.SubnetMask != "255.255.248.0" || profile.IPv4.Gateway != "172.16.40.2" {
		t.Fatalf("IPv4 = %+v", profile.IPv4)
	}
	if !reflect.DeepEqual(profile.DNS, []string{"172.16.40.2"}) {
		t.Fatalf("DNS = %v", profile.DNS)
	}
}

func TestReadProfileSupportsCurrentAddressAndSSID(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "blacklake-proxy", currentAddressProfileYAML)

	profile, err := ReadProfile(dir, "blacklake-proxy")
	if err != nil {
		t.Fatal(err)
	}
	if !profile.CurrentAddress || profile.IPv4.Address != "" {
		t.Fatalf("address mode = %+v", profile)
	}
	if profile.SSID != "blacklake" {
		t.Fatalf("SSID = %q", profile.SSID)
	}
	if profile.IPv4.SubnetMask != "255.255.248.0" || profile.IPv4.Gateway != "172.16.40.2" {
		t.Fatalf("IPv4 = %+v", profile.IPv4)
	}
}

func TestUserProfileOverridesBuiltin(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "blacklake-proxy", validProfileYAML)

	profile, err := ReadProfile(dir, "blacklake-proxy")
	if err != nil {
		t.Fatal(err)
	}
	if profile.CurrentAddress || profile.IPv4.Address != "172.16.40.100" {
		t.Fatalf("profile = %+v", profile)
	}
}

func TestReadProfileRejectsUnknownField(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "bad", validProfileYAML+"unknown: true\n")

	_, err := ReadProfile(dir, "bad")
	if err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("err = %v", err)
	}
}

func TestReadProfileRejectsUnsafeName(t *testing.T) {
	_, err := ReadProfile(t.TempDir(), "../profile")
	if err == nil || !strings.Contains(err.Error(), "无效") {
		t.Fatalf("err = %v", err)
	}
}

func TestReadProfileValidatesNetworkValues(t *testing.T) {
	dir := t.TempDir()
	bad := strings.Replace(validProfileYAML, "255.255.255.0", "255.0.255.0", 1)
	writeProfile(t, dir, "bad-mask", bad)

	_, err := ReadProfile(dir, "bad-mask")
	if err == nil || !strings.Contains(err.Error(), "subnet_mask") {
		t.Fatalf("err = %v", err)
	}
}

func TestReadProfileRejectsUnknownAddressMode(t *testing.T) {
	dir := t.TempDir()
	bad := strings.Replace(currentAddressProfileYAML, "address: current", "address: dynamic", 1)
	writeProfile(t, dir, "bad-address", bad)

	_, err := ReadProfile(dir, "bad-address")
	if err == nil || !strings.Contains(err.Error(), "ipv4.address") {
		t.Fatalf("err = %v", err)
	}
}

func TestListProfilesSorted(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "test-network", validProfileYAML)
	writeProfile(t, dir, "blacklake-proxy", validProfileYAML)
	if err := os.WriteFile(filepath.Join(dir, ".active.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := ListProfiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"blacklake-proxy", "test-network"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("profiles = %v, want %v", got, want)
	}
}

func TestListProfilesIncludesBuiltinWithoutConfigDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	got, err := ListProfiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"blacklake-proxy"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("profiles = %v, want %v", got, want)
	}
}

func TestActiveStateRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "network")
	want := ActiveState{
		Profile:   "blacklake-proxy",
		Interface: "wifi",
		Device:    "en0",
		Service:   "Office Wireless",
		IPv4: IPv4Config{
			Address:    "172.16.40.100",
			SubnetMask: "255.255.255.0",
			Gateway:    "172.16.40.2",
		},
		DNS: []string{"172.16.40.2"},
	}
	if err := WriteActive(dir, want); err != nil {
		t.Fatal(err)
	}
	got, active, err := ReadActive(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !active || got.Profile != want.Profile || got.Device != want.Device {
		t.Fatalf("state = %+v, active = %v", got, active)
	}
	if err := RemoveActive(dir); err != nil {
		t.Fatal(err)
	}
	if _, active, err := ReadActive(dir); err != nil || active {
		t.Fatalf("active after remove = %v, err = %v", active, err)
	}
}
