package tunnel

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	calls      []string
	running    bool
	startError error
	exitError  error
}

func (f *fakeRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, commandKey(name, args...))
	if f.startError != nil {
		return f.startError
	}
	f.running = true
	return nil
}

func (f *fakeRunner) Output(name string, args ...string) (string, error) {
	f.calls = append(f.calls, commandKey(name, args...))
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "-O check") {
		if f.running {
			return "Master running (pid=123)", nil
		}
		return "", errors.New("not running")
	}
	if strings.Contains(joined, "-O exit") {
		if f.exitError != nil {
			return "", f.exitError
		}
		f.running = false
		return "Exit request sent.", nil
	}
	return "", errors.New("unexpected command")
}

func commandKey(name string, args ...string) string {
	return strings.Join(append([]string{name}, args...), " ")
}

func newTestManager(t *testing.T) (*Manager, *fakeRunner) {
	t.Helper()
	configDir := filepath.Join(t.TempDir(), "tunnels")
	runtimeDir := filepath.Join(t.TempDir(), "run")
	writeProfile(t, configDir, "studio", validProfileYAML)
	runner := &fakeRunner{}
	manager := &Manager{
		ConfigDir:  configDir,
		RuntimeDir: runtimeDir,
		ssh:        "ssh",
		runner:     runner,
		available:  func(int) error { return nil },
		reachable:  func(int) bool { return runner.running },
	}
	return manager, runner
}

func TestUpStartsControlMasterAndIsIdempotent(t *testing.T) {
	manager, runner := newTestManager(t)

	result, err := manager.Up("studio")
	if err != nil {
		t.Fatal(err)
	}
	if result.AlreadyRunning || !result.Status.Running {
		t.Fatalf("result = %+v", result)
	}
	if mode := fileMode(t, manager.RuntimeDir); mode != 0o700 {
		t.Fatalf("runtime mode = %o", mode)
	}
	start := findCall(runner.calls, "-M -N -f -T")
	for _, want := range []string{
		"-S " + manager.socketPath("studio"),
		"-o ServerAliveInterval=30",
		"-o ExitOnForwardFailure=yes",
		"-L 127.0.0.1:19010:127.0.0.1:9010",
		"-L 127.0.0.1:19119:[::1]:9119",
		"mac-studio",
	} {
		if !strings.Contains(start, want) {
			t.Fatalf("start command missing %q: %s", want, start)
		}
	}

	startCalls := countCalls(runner.calls, "-M -N -f -T")
	result, err = manager.Up("studio")
	if err != nil {
		t.Fatal(err)
	}
	if !result.AlreadyRunning {
		t.Fatal("expected already running")
	}
	if got := countCalls(runner.calls, "-M -N -f -T"); got != startCalls {
		t.Fatalf("start calls = %d, want %d", got, startCalls)
	}
}

func TestUpRejectsOccupiedPortBeforeSSH(t *testing.T) {
	manager, runner := newTestManager(t)
	manager.available = func(port int) error {
		if port == 19010 {
			return errors.New("address already in use")
		}
		return nil
	}

	_, err := manager.Up("studio")
	if err == nil || !strings.Contains(err.Error(), "本地端口 19010 不可用") {
		t.Fatalf("err = %v", err)
	}
	if countCalls(runner.calls, "-M -N -f -T") != 0 {
		t.Fatalf("unexpected start: %v", runner.calls)
	}
}

func TestDownStopsOnlyProfileControlMaster(t *testing.T) {
	manager, runner := newTestManager(t)
	if _, err := manager.Up("studio"); err != nil {
		t.Fatal(err)
	}

	result, err := manager.Down("studio")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Stopped || runner.running {
		t.Fatalf("result = %+v, running = %v", result, runner.running)
	}
	exit := findCall(runner.calls, "-O exit")
	if !strings.Contains(exit, "-S "+manager.socketPath("studio")) || !strings.HasSuffix(exit, "mac-studio") {
		t.Fatalf("exit command = %s", exit)
	}

	result, err = manager.Down("studio")
	if err != nil || result.Stopped {
		t.Fatalf("second down = %+v, err = %v", result, err)
	}
}

func TestUpRemovesStaleSocket(t *testing.T) {
	manager, _ := newTestManager(t)
	if err := os.MkdirAll(manager.RuntimeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	socket := manager.socketPath("studio")
	if err := os.WriteFile(socket, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.Up("studio"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(socket); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale socket still exists: %v", err)
	}
}

func fileMode(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}

func findCall(calls []string, part string) string {
	for _, call := range calls {
		if strings.Contains(call, part) {
			return call
		}
	}
	return ""
}

func countCalls(calls []string, part string) int {
	count := 0
	for _, call := range calls {
		if strings.Contains(call, part) {
			count++
		}
	}
	return count
}
