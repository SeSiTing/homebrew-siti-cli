package shell

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// Run with -update to refresh golden files after intentional wrapper changes:
//
//	go test ./internal/shell -update
var update = flag.Bool("update", false, "update wrapper snapshot golden files")

func TestWrapper_Snapshot(t *testing.T) {
	cases := []struct {
		shell  string
		golden string
	}{
		{"zsh", "wrapper_zsh.golden"},
		{"bash", "wrapper_zsh.golden"}, // bash uses the posix wrapper too
		{"sh", "wrapper_zsh.golden"},
		{"fish", "wrapper_fish.golden"},
	}
	for _, tc := range cases {
		t.Run(tc.shell, func(t *testing.T) {
			got := WrapperFor(tc.shell)
			path := filepath.Join("testdata", tc.golden)

			if *update {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden %s: %v (run with -update to create)", path, err)
			}
			if got != string(want) {
				t.Errorf("WrapperFor(%q) does not match %s\nrun: go test ./internal/shell -update", tc.shell, path)
			}
		})
	}
}

// TestWrapper_NoGrepFilter ensures the wrapper does not contain a grep
// allowlist — past iterations had `grep -E '^(export |unset |if )'` which
// silently dropped any new shell helper not in the list.
func TestWrapper_NoGrepFilter(t *testing.T) {
	for _, sh := range []string{"zsh", "fish"} {
		t.Run(sh, func(t *testing.T) {
			w := WrapperFor(sh)
			if contains(w, "grep -E") {
				t.Errorf("%s wrapper contains grep filter — must trust binary stdout", sh)
			}
		})
	}
}

func TestWrapper_GrokModelInjection(t *testing.T) {
	for _, sh := range []string{"zsh", "fish"} {
		t.Run(sh, func(t *testing.T) {
			w := WrapperFor(sh)
			if !contains(w, "SITI_GROK_MODEL_ID") || !contains(w, "--model") {
				t.Errorf("%s wrapper does not inject the selected Grok model", sh)
			}
		})
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
