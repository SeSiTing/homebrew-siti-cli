package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	profileVersion = 1
	activeFileName = ".active.json"
)

var profileNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

var builtinProfiles = map[string]Profile{
	"blacklake-proxy": {
		Version:           profileVersion,
		Interface:         "wifi",
		SSID:              "blacklake",
		CurrentAddress:    true,
		CurrentSubnetMask: true,
		IPv4: IPv4Config{
			Gateway: "172.16.40.2",
		},
		DNS: []string{"172.16.40.2"},
	},
}

type IPv4Config struct {
	Address    string `yaml:"address" json:"address"`
	SubnetMask string `yaml:"subnet_mask" json:"subnet_mask"`
	Gateway    string `yaml:"gateway" json:"gateway"`
}

type Profile struct {
	Version           int        `yaml:"version"`
	Interface         string     `yaml:"interface"`
	SSID              string     `yaml:"ssid,omitempty"`
	CurrentAddress    bool       `yaml:"-"`
	CurrentSubnetMask bool       `yaml:"-"`
	IPv4              IPv4Config `yaml:"ipv4"`
	DNS               []string   `yaml:"dns"`
}

type ActiveState struct {
	Version   int        `json:"version"`
	Profile   string     `json:"profile"`
	Interface string     `json:"interface"`
	SSID      string     `json:"ssid,omitempty"`
	Device    string     `json:"device"`
	Service   string     `json:"service"`
	IPv4      IPv4Config `json:"ipv4"`
	DNS       []string   `json:"dns"`
}

func DefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录: %w", err)
	}
	return filepath.Join(home, ".siti", "network"), nil
}

func ReadProfile(dir, name string) (Profile, error) {
	if !profileNamePattern.MatchString(name) {
		return Profile{}, fmt.Errorf("无效的 network profile 名称 %q", name)
	}

	path := filepath.Join(dir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if profile, ok := builtinProfile(name); ok {
				return profile, nil
			}
			return Profile{}, fmt.Errorf("network profile %q 不存在: %s", name, path)
		}
		return Profile{}, fmt.Errorf("读取 network profile %q: %w", name, err)
	}

	var profile Profile
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&profile); err != nil {
		return Profile{}, fmt.Errorf("解析 network profile %q: %w", name, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Profile{}, fmt.Errorf("network profile %q 只能包含一个 YAML 文档", name)
		}
		return Profile{}, fmt.Errorf("解析 network profile %q: %w", name, err)
	}

	profile.Interface = strings.ToLower(strings.TrimSpace(profile.Interface))
	profile.SSID = strings.TrimSpace(profile.SSID)
	profile.IPv4.Address = strings.TrimSpace(profile.IPv4.Address)
	profile.IPv4.SubnetMask = strings.TrimSpace(profile.IPv4.SubnetMask)
	profile.IPv4.Gateway = strings.TrimSpace(profile.IPv4.Gateway)
	if strings.EqualFold(profile.IPv4.Address, "current") {
		profile.CurrentAddress = true
		profile.IPv4.Address = ""
	}
	if strings.EqualFold(profile.IPv4.SubnetMask, "current") {
		profile.CurrentSubnetMask = true
		profile.IPv4.SubnetMask = ""
	}
	for i := range profile.DNS {
		profile.DNS[i] = strings.TrimSpace(profile.DNS[i])
	}
	if err := profile.Validate(); err != nil {
		return Profile{}, fmt.Errorf("network profile %q: %w", name, err)
	}
	return profile, nil
}

func (p Profile) Validate() error {
	if p.Version != profileVersion {
		return fmt.Errorf("version 必须为 %d", profileVersion)
	}
	if p.Interface != "wifi" {
		return fmt.Errorf("interface 目前仅支持 wifi")
	}
	if !p.CurrentAddress && !isIPv4(p.IPv4.Address) {
		return fmt.Errorf("ipv4.address 不是有效 IPv4 地址: %q", p.IPv4.Address)
	}
	if !p.CurrentSubnetMask && !isSubnetMask(p.IPv4.SubnetMask) {
		return fmt.Errorf("ipv4.subnet_mask 不是有效子网掩码: %q", p.IPv4.SubnetMask)
	}
	if !isIPv4(p.IPv4.Gateway) {
		return fmt.Errorf("ipv4.gateway 不是有效 IPv4 地址: %q", p.IPv4.Gateway)
	}
	if len(p.DNS) == 0 {
		return fmt.Errorf("dns 至少需要一个服务器地址")
	}
	for _, server := range p.DNS {
		if net.ParseIP(server) == nil {
			return fmt.Errorf("dns 包含无效地址: %q", server)
		}
	}
	return nil
}

func ListProfiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return builtinProfileNames(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取 network profile 目录: %w", err)
	}

	profiles := builtinProfileNames()
	seen := make(map[string]bool, len(profiles)+len(entries))
	for _, name := range profiles {
		seen[name] = true
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		if profileNamePattern.MatchString(name) && !seen[name] {
			profiles = append(profiles, name)
			seen[name] = true
		}
	}
	sort.Strings(profiles)
	return profiles, nil
}

func builtinProfile(name string) (Profile, bool) {
	profile, ok := builtinProfiles[name]
	if !ok {
		return Profile{}, false
	}
	profile.DNS = append([]string(nil), profile.DNS...)
	return profile, true
}

func builtinProfileNames() []string {
	names := make([]string, 0, len(builtinProfiles))
	for name := range builtinProfiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func ReadActive(dir string) (ActiveState, bool, error) {
	data, err := os.ReadFile(filepath.Join(dir, activeFileName))
	if errors.Is(err, os.ErrNotExist) {
		return ActiveState{}, false, nil
	}
	if err != nil {
		return ActiveState{}, false, fmt.Errorf("读取 network active 状态: %w", err)
	}

	var state ActiveState
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return ActiveState{}, false, fmt.Errorf("解析 network active 状态: %w", err)
	}
	if state.Version != profileVersion || state.Profile == "" || state.Device == "" || state.Service == "" {
		return ActiveState{}, false, fmt.Errorf("network active 状态不完整")
	}
	return state, true, nil
}

func WriteActive(dir string, state ActiveState) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("创建 network profile 目录: %w", err)
	}
	state.Version = profileVersion
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("编码 network active 状态: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, ".active-*")
	if err != nil {
		return fmt.Errorf("创建 network active 临时文件: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("写入 network active 状态: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭 network active 状态: %w", err)
	}
	if err := os.Rename(tmpPath, filepath.Join(dir, activeFileName)); err != nil {
		return fmt.Errorf("保存 network active 状态: %w", err)
	}
	return nil
}

func RemoveActive(dir string) error {
	err := os.Remove(filepath.Join(dir, activeFileName))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("删除 network active 状态: %w", err)
	}
	return nil
}

func isIPv4(value string) bool {
	ip := net.ParseIP(value)
	return ip != nil && ip.To4() != nil
}

func isSubnetMask(value string) bool {
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() == nil {
		return false
	}
	ones, bits := net.IPMask(ip.To4()).Size()
	return bits == 32 && ones > 0
}
