package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindBlOpsEditableProjectFromSrcPth(t *testing.T) {
	tmp := t.TempDir()
	project := filepath.Join(tmp, "workspace", "share", "bl-ops")
	if err := os.MkdirAll(filepath.Join(project, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "pyproject.toml"), []byte("[project]\nname = \"bl-ops\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sitePackages := filepath.Join(tmp, "tools", "bl-ops", "lib", "python3.13", "site-packages")
	if err := os.MkdirAll(sitePackages, 0o755); err != nil {
		t.Fatal(err)
	}
	pth := filepath.Join(sitePackages, "_editable_impl_bl_ops.pth")
	if err := os.WriteFile(pth, []byte(filepath.Join(project, "src")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, ok := findBlOpsEditableProject(filepath.Join(tmp, "tools"))
	if !ok {
		t.Fatal("expected editable project")
	}
	if got != project {
		t.Fatalf("project = %q, want %q", got, project)
	}
}

func TestFindBlOpsEditableProjectMissing(t *testing.T) {
	got, ok := findBlOpsEditableProject(t.TempDir())
	if ok {
		t.Fatalf("project = %q, want missing", got)
	}
}
