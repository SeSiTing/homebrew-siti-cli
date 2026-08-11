package tunnel

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const profileVersion = 1

var profileNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

var builtinProfiles = map[string]Profile{
	"mac-studio": {
		Version: 1,
		Target:  "mac-studio",
		Forwards: []Forward{
			{Name: "openclaw", LocalPort: 19010, RemoteHost: "127.0.0.1", RemotePort: 9010, URL: "http://127.0.0.1:19010/"},
			{Name: "hermes", LocalPort: 19119, RemoteHost: "127.0.0.1", RemotePort: 9119, URL: "http://127.0.0.1:19119/"},
		},
	},
}

type Forward struct {
	Name       string `yaml:"name"`
	LocalPort  int    `yaml:"local_port"`
	RemoteHost string `yaml:"remote_host,omitempty"`
	RemotePort int    `yaml:"remote_port"`
	URL        string `yaml:"url,omitempty"`
}

type Profile struct {
	Version  int       `yaml:"version"`
	Target   string    `yaml:"target"`
	Forwards []Forward `yaml:"forwards"`
}

func DefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录: %w", err)
	}
	return filepath.Join(home, ".siti", "tunnels"), nil
}

func DefaultRuntimeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录: %w", err)
	}
	return filepath.Join(home, ".siti", "run"), nil
}

func ReadProfile(dir, name string) (Profile, error) {
	if !profileNamePattern.MatchString(name) {
		return Profile{}, fmt.Errorf("无效的 tunnel profile 名称 %q", name)
	}

	path := filepath.Join(dir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if profile, ok := builtinProfile(name); ok {
				return profile, nil
			}
			return Profile{}, fmt.Errorf("tunnel profile %q 不存在: %s", name, path)
		}
		return Profile{}, fmt.Errorf("读取 tunnel profile %q: %w", name, err)
	}

	var profile Profile
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&profile); err != nil {
		return Profile{}, fmt.Errorf("解析 tunnel profile %q: %w", name, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Profile{}, fmt.Errorf("tunnel profile %q 只能包含一个 YAML 文档", name)
		}
		return Profile{}, fmt.Errorf("解析 tunnel profile %q: %w", name, err)
	}

	profile.Target = strings.TrimSpace(profile.Target)
	for i := range profile.Forwards {
		profile.Forwards[i].Name = strings.TrimSpace(profile.Forwards[i].Name)
		profile.Forwards[i].RemoteHost = strings.TrimSpace(profile.Forwards[i].RemoteHost)
		profile.Forwards[i].URL = strings.TrimSpace(profile.Forwards[i].URL)
		if profile.Forwards[i].RemoteHost == "" {
			profile.Forwards[i].RemoteHost = "127.0.0.1"
		}
	}
	if err := profile.Validate(); err != nil {
		return Profile{}, fmt.Errorf("tunnel profile %q: %w", name, err)
	}
	return profile, nil
}

func (p Profile) Validate() error {
	if p.Version != profileVersion {
		return fmt.Errorf("version 必须为 %d", profileVersion)
	}
	if p.Target == "" || strings.HasPrefix(p.Target, "-") || hasWhitespaceOrControl(p.Target) {
		return fmt.Errorf("target 不是有效的 SSH 目标: %q", p.Target)
	}
	if len(p.Forwards) == 0 {
		return fmt.Errorf("forwards 至少需要一项")
	}

	names := make(map[string]bool, len(p.Forwards))
	ports := make(map[int]bool, len(p.Forwards))
	for i, forward := range p.Forwards {
		field := fmt.Sprintf("forwards[%d]", i)
		if !profileNamePattern.MatchString(forward.Name) {
			return fmt.Errorf("%s.name 无效: %q", field, forward.Name)
		}
		if names[forward.Name] {
			return fmt.Errorf("forward 名称重复: %q", forward.Name)
		}
		names[forward.Name] = true
		if !validPort(forward.LocalPort) {
			return fmt.Errorf("%s.local_port 必须在 1-65535 之间", field)
		}
		if ports[forward.LocalPort] {
			return fmt.Errorf("local_port 重复: %d", forward.LocalPort)
		}
		ports[forward.LocalPort] = true
		if !validRemoteHost(forward.RemoteHost) {
			return fmt.Errorf("%s.remote_host 无效: %q", field, forward.RemoteHost)
		}
		if !validPort(forward.RemotePort) {
			return fmt.Errorf("%s.remote_port 必须在 1-65535 之间", field)
		}
		if forward.URL != "" {
			if err := validateURL(forward.URL, forward.LocalPort); err != nil {
				return fmt.Errorf("%s.url: %w", field, err)
			}
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
		return nil, fmt.Errorf("读取 tunnel profile 目录: %w", err)
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
	profile.Forwards = append([]Forward(nil), profile.Forwards...)
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

func hasWhitespaceOrControl(value string) bool {
	return strings.IndexFunc(value, func(r rune) bool {
		return r <= ' ' || r == 0x7f
	}) >= 0
}

func validPort(port int) bool {
	return port >= 1 && port <= 65535
}

func validRemoteHost(host string) bool {
	if host == "" || strings.HasPrefix(host, "-") || hasWhitespaceOrControl(host) {
		return false
	}
	if strings.ContainsAny(host, "[]") {
		return false
	}
	return !strings.Contains(host, ":") || net.ParseIP(host) != nil
}

func validateURL(value string, localPort int) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("不是有效 URL: %q", value)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("scheme 仅支持 http 或 https")
	}
	if parsed.User != nil || parsed.RawQuery != "" {
		return fmt.Errorf("URL 不允许包含凭证或 query 参数")
	}
	host := parsed.Hostname()
	if host != "localhost" {
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			return fmt.Errorf("host 必须是本机 loopback")
		}
	}
	port := parsed.Port()
	if port == "" {
		if (parsed.Scheme == "http" && localPort != 80) || (parsed.Scheme == "https" && localPort != 443) {
			return fmt.Errorf("URL 端口必须与 local_port %d 一致", localPort)
		}
		return nil
	}
	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort != localPort {
		return fmt.Errorf("URL 端口必须与 local_port %d 一致", localPort)
	}
	return nil
}
