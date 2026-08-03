package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrewFormulaOutdated(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		want        bool
		wantErr     bool
		errContains string
	}{
		{
			name:   "outdated",
			script: "printf '%s\\n' 'sesiting/tap/siti-cli (2.0.20) < 2.0.21'\nexit 1\n",
			want:   true,
		},
		{
			name:   "current",
			script: "exit 0\n",
		},
		{
			name:        "brew error",
			script:      "printf '%s\\n' 'Refusing to load formula from untrusted tap' >&2\nexit 1\n",
			wantErr:     true,
			errContains: "untrusted tap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			brew := filepath.Join(dir, "brew")
			if err := os.WriteFile(brew, []byte("#!/bin/sh\n"+tt.script), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir)

			got, err := brewFormulaOutdated(sitiBrewFormula)
			if got != tt.want {
				t.Fatalf("brewFormulaOutdated() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("brewFormulaOutdated() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("brewFormulaOutdated() error = %q, want substring %q", err, tt.errContains)
			}
		})
	}
}
