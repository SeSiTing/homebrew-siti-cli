package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const blOpsGitSpec = "git+ssh://git@gitlab.blacklake.tech/harness/bl-ops.git"

type blOpsInstall struct {
	Installed       bool
	ExecutablePath  string
	EditableProject string
	ToolsDir        string
}

func sectionBlOps() error {
	if upgradeDryRun {
		return sectionBlOpsDryRun()
	}

	fmt.Println("── bl-ops ──")
	if _, err := exec.LookPath("uv"); err != nil {
		fmt.Println("✗ uv 未安装")
		return nil
	}

	install := inspectBlOpsInstall()
	if install.EditableProject != "" {
		fmt.Printf("! editable 源: %s\n", install.EditableProject)
		fmt.Println("→ uv sync")
		if err := runCmd("uv", "sync", "--project", install.EditableProject); err != nil {
			return fmt.Errorf("uv sync 失败: %w", err)
		}
		fmt.Println("→ uv tool install --editable")
		if err := runCmd("uv", "tool", "install", "--editable", install.EditableProject, "--force"); err != nil {
			return fmt.Errorf("editable 安装失败: %w", err)
		}
		fmt.Println("✓ bl-ops 已从本地源码刷新")
		return nil
	}

	if install.Installed {
		fmt.Println("→ uv tool upgrade bl-ops")
		if err := runCmd("uv", "tool", "upgrade", "bl-ops"); err == nil {
			fmt.Println("✓ bl-ops 已升级")
			return nil
		}
		fmt.Println("! uv tool upgrade 失败，改用 git 源强制安装")
	} else {
		fmt.Println("! 未检测到 bl-ops，开始安装")
	}

	if err := runCmd("uv", "tool", "install", blOpsGitSpec, "--force", "--upgrade"); err != nil {
		return fmt.Errorf("安装失败: %w", err)
	}
	fmt.Println("✓ bl-ops 已安装/升级")
	return nil
}

func sectionBlOpsDryRun() error {
	fmt.Println("── bl-ops ── (dry-run)")
	if _, err := exec.LookPath("uv"); err != nil {
		fmt.Println("✗ uv 未安装")
		return nil
	}

	install := inspectBlOpsInstall()
	if install.ExecutablePath != "" {
		fmt.Printf("  executable: %s\n", install.ExecutablePath)
	}
	if install.ToolsDir != "" {
		fmt.Printf("  uv tools: %s\n", install.ToolsDir)
	}

	if install.EditableProject != "" {
		fmt.Printf("  editable 源: %s\n", install.EditableProject)
		fmt.Println("  would run:")
		fmt.Printf("    uv sync --project %s\n", install.EditableProject)
		fmt.Printf("    uv tool install --editable %s --force\n", install.EditableProject)
		return nil
	}

	if install.Installed {
		fmt.Println("  would run:")
		fmt.Println("    uv tool upgrade bl-ops")
		return nil
	}

	fmt.Println("  bl-ops 未安装")
	fmt.Println("  would run:")
	fmt.Printf("    uv tool install %s --force --upgrade\n", blOpsGitSpec)
	return nil
}

func inspectBlOpsInstall() blOpsInstall {
	var install blOpsInstall
	if path, err := exec.LookPath("bl-ops"); err == nil {
		install.Installed = true
		install.ExecutablePath = path
	}

	toolsDir, ok := uvToolDir()
	if !ok {
		return install
	}
	install.ToolsDir = toolsDir

	if project, ok := findBlOpsEditableProject(toolsDir); ok {
		install.EditableProject = project
	}

	return install
}

func uvToolDir() (string, bool) {
	var buf bytes.Buffer
	if err := runCmdOutput(&buf, "uv", "tool", "dir"); err != nil {
		return "", false
	}
	dir := strings.TrimSpace(buf.String())
	if dir == "" {
		return "", false
	}
	return dir, true
}

func findBlOpsEditableProject(toolsDir string) (string, bool) {
	pattern := filepath.Join(toolsDir, "bl-ops", "lib", "python*", "site-packages", "_editable*_bl_ops.pth")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", false
	}

	for _, match := range matches {
		project, ok := editableProjectFromPth(match)
		if ok {
			return project, true
		}
	}
	return "", false
}

func editableProjectFromPth(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	for _, line := range strings.Split(string(data), "\n") {
		source := strings.TrimSpace(line)
		if source == "" || strings.HasPrefix(source, "import ") {
			continue
		}

		if strings.HasSuffix(source, string(filepath.Separator)+"src") {
			candidate := filepath.Dir(source)
			if fileExists(filepath.Join(candidate, "pyproject.toml")) {
				return candidate, true
			}
		}

		if fileExists(filepath.Join(source, "pyproject.toml")) {
			return source, true
		}
	}

	return "", false
}
