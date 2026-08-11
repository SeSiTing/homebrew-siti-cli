package tunnel

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	return strings.TrimSpace(string(out)), nil
}

type Manager struct {
	ConfigDir  string
	RuntimeDir string
	ssh        string
	runner     commandRunner
	available  func(int) error
	reachable  func(int) bool
}

type ForwardStatus struct {
	Forward
	Reachable bool
}

type StatusResult struct {
	Profile  Profile
	Running  bool
	Detail   string
	Forwards []ForwardStatus
}

type UpResult struct {
	Status         StatusResult
	AlreadyRunning bool
}

type DownResult struct {
	Stopped bool
}

func NewManager() (*Manager, error) {
	configDir, err := DefaultConfigDir()
	if err != nil {
		return nil, err
	}
	runtimeDir, err := DefaultRuntimeDir()
	if err != nil {
		return nil, err
	}
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		return nil, fmt.Errorf("未找到 ssh: %w", err)
	}
	return &Manager{
		ConfigDir:  configDir,
		RuntimeDir: runtimeDir,
		ssh:        ssh,
		runner:     execCommandRunner{},
		available:  localPortAvailable,
		reachable:  localPortReachable,
	}, nil
}

func (m *Manager) Up(name string) (UpResult, error) {
	profile, err := ReadProfile(m.ConfigDir, name)
	if err != nil {
		return UpResult{}, err
	}
	current := m.status(name, profile)
	if current.Running {
		return UpResult{Status: current, AlreadyRunning: true}, nil
	}

	for _, forward := range profile.Forwards {
		if err := m.available(forward.LocalPort); err != nil {
			return UpResult{}, fmt.Errorf("本地端口 %d 不可用: %w", forward.LocalPort, err)
		}
	}
	if err := os.MkdirAll(m.RuntimeDir, 0o700); err != nil {
		return UpResult{}, fmt.Errorf("创建 tunnel runtime 目录: %w", err)
	}
	if err := os.Chmod(m.RuntimeDir, 0o700); err != nil {
		return UpResult{}, fmt.Errorf("保护 tunnel runtime 目录: %w", err)
	}
	socket := m.socketPath(name)
	if err := os.Remove(socket); err != nil && !errors.Is(err, os.ErrNotExist) {
		return UpResult{}, fmt.Errorf("清理旧 tunnel socket: %w", err)
	}

	args := []string{
		"-M", "-N", "-f", "-T",
		"-S", socket,
		"-o", "ControlMaster=yes",
		"-o", "ControlPersist=no",
		"-o", "ServerAliveInterval=30",
		"-o", "ServerAliveCountMax=3",
		"-o", "ExitOnForwardFailure=yes",
	}
	for _, forward := range profile.Forwards {
		args = append(args, "-L", forwardSpec(forward))
	}
	args = append(args, profile.Target)
	if err := m.runner.Run(m.ssh, args...); err != nil {
		_ = os.Remove(socket)
		return UpResult{}, fmt.Errorf("启动 SSH tunnel: %w", err)
	}

	result := m.status(name, profile)
	if !result.Running {
		_, _ = m.runner.Output(m.ssh, "-S", socket, "-O", "exit", profile.Target)
		_ = os.Remove(socket)
		return UpResult{}, fmt.Errorf("SSH tunnel 启动后未进入运行状态: %s", result.Detail)
	}
	return UpResult{Status: result}, nil
}

func (m *Manager) Down(name string) (DownResult, error) {
	profile, err := ReadProfile(m.ConfigDir, name)
	if err != nil {
		return DownResult{}, err
	}
	current := m.status(name, profile)
	if !current.Running {
		_ = os.Remove(m.socketPath(name))
		return DownResult{}, nil
	}
	if _, err := m.runner.Output(m.ssh, "-S", m.socketPath(name), "-O", "exit", profile.Target); err != nil {
		return DownResult{}, fmt.Errorf("关闭 SSH tunnel: %w", err)
	}
	_ = os.Remove(m.socketPath(name))
	return DownResult{Stopped: true}, nil
}

func (m *Manager) Status(name string) (StatusResult, error) {
	profile, err := ReadProfile(m.ConfigDir, name)
	if err != nil {
		return StatusResult{}, err
	}
	return m.status(name, profile), nil
}

func (m *Manager) List() ([]string, error) {
	return ListProfiles(m.ConfigDir)
}

func (m *Manager) status(name string, profile Profile) StatusResult {
	detail, err := m.runner.Output(m.ssh, "-S", m.socketPath(name), "-O", "check", profile.Target)
	result := StatusResult{
		Profile:  profile,
		Running:  err == nil,
		Detail:   detail,
		Forwards: make([]ForwardStatus, 0, len(profile.Forwards)),
	}
	if err != nil {
		result.Detail = "SSH master 未运行"
	}
	for _, forward := range profile.Forwards {
		result.Forwards = append(result.Forwards, ForwardStatus{
			Forward:   forward,
			Reachable: m.reachable(forward.LocalPort),
		})
	}
	return result
}

func (m *Manager) socketPath(name string) string {
	return filepath.Join(m.RuntimeDir, "tunnel-"+name+".sock")
}

func forwardSpec(forward Forward) string {
	host := forward.RemoteHost
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return "127.0.0.1:" + strconv.Itoa(forward.LocalPort) + ":" + host + ":" + strconv.Itoa(forward.RemotePort)
}

func localPortAvailable(port int) error {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return err
	}
	return listener.Close()
}

func localPortReachable(port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), 300*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
