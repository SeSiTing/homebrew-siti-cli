package tunnel

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const validProfileYAML = `version: 1
target: mac-studio
forwards:
  - name: openclaw
    local_port: 19010
    remote_port: 9010
    url: http://127.0.0.1:19010/
  - name: hermes
    local_port: 19119
    remote_host: ::1
    remote_port: 9119
    url: http://localhost:19119/
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
	writeProfile(t, dir, "studio", validProfileYAML)

	profile, err := ReadProfile(dir, "studio")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Target != "mac-studio" || len(profile.Forwards) != 2 {
		t.Fatalf("profile = %+v", profile)
	}
	if profile.Forwards[0].RemoteHost != "127.0.0.1" {
		t.Fatalf("default remote host = %q", profile.Forwards[0].RemoteHost)
	}
}

func TestReadBuiltinProfileUsesSSHHostAlias(t *testing.T) {
	profile, err := ReadProfile(t.TempDir(), "mac-studio")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Target != "mac-studio" || len(profile.Forwards) != 2 {
		t.Fatalf("profile = %+v", profile)
	}
	if profile.Forwards[0].LocalPort != 19010 || profile.Forwards[0].RemotePort != 9010 {
		t.Fatalf("openclaw forward = %+v", profile.Forwards[0])
	}
}

func TestUserProfileOverridesBuiltin(t *testing.T) {
	dir := t.TempDir()
	custom := strings.Replace(validProfileYAML, "target: mac-studio", "target: studio-vpn", 1)
	writeProfile(t, dir, "mac-studio", custom)

	profile, err := ReadProfile(dir, "mac-studio")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Target != "studio-vpn" {
		t.Fatalf("target = %q", profile.Target)
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
	_, err := ReadProfile(t.TempDir(), "../studio")
	if err == nil || !strings.Contains(err.Error(), "无效") {
		t.Fatalf("err = %v", err)
	}
}

func TestProfileValidation(t *testing.T) {
	tests := []struct {
		name    string
		replace string
		with    string
		want    string
	}{
		{name: "target option injection", replace: "target: mac-studio", with: "target: -Fbad", want: "SSH 目标"},
		{name: "duplicate port", replace: "local_port: 19119", with: "local_port: 19010", want: "local_port 重复"},
		{name: "public URL", replace: "http://localhost:19119/", with: "http://example.com:19119/", want: "loopback"},
		{name: "mismatched URL port", replace: "http://127.0.0.1:19010/", with: "http://127.0.0.1:19011/", want: "必须与 local_port"},
		{name: "URL credentials", replace: "http://127.0.0.1:19010/", with: "http://user:secret@127.0.0.1:19010/", want: "不允许包含凭证"},
		{name: "URL query", replace: "http://127.0.0.1:19010/", with: "http://127.0.0.1:19010/?token=secret", want: "不允许包含凭证"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			content := strings.Replace(validProfileYAML, tt.replace, tt.with, 1)
			writeProfile(t, dir, "bad", content)
			_, err := ReadProfile(dir, "bad")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestListProfilesSorted(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "studio-z", validProfileYAML)
	writeProfile(t, dir, "studio-a", validProfileYAML)
	if err := os.WriteFile(filepath.Join(dir, "ignore.txt"), []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := ListProfiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"mac-studio", "studio-a", "studio-z"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("profiles = %v, want %v", got, want)
	}
}
