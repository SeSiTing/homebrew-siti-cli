package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

const sitiBrewFormula = "SeSiTing/tap/siti-cli"

// sectionSelf upgrades siti-cli. Returns true if an update was applied.
func sectionSelf(cmd *cobra.Command) (bool, error) {
	fmt.Println("── siti-cli ──")
	fmt.Printf("当前版本: v%s\n", cmd.Root().Version)

	installMethod := os.Getenv("INSTALL_METHOD")
	if installMethod == "" {
		if _, err := exec.LookPath("brew"); err == nil {
			out, _ := exec.Command("brew", "list", "--formula", "siti-cli").Output()
			if len(out) > 0 {
				installMethod = "homebrew"
			}
		}
	}

	switch installMethod {
	case "homebrew", "":
		if _, err := exec.LookPath("brew"); err != nil {
			return false, fmt.Errorf("未找到 Homebrew: %w", err)
		}
		if err := runCmd("brew", "update"); err != nil {
			return false, fmt.Errorf("brew update 失败: %w", err)
		}
		outdated, err := brewFormulaOutdated(sitiBrewFormula)
		if err != nil {
			return false, fmt.Errorf("检查 siti-cli 更新失败: %w", err)
		}
		if !outdated {
			fmt.Println("✓ 已是最新版本")
			return false, nil
		}
		fmt.Printf("→ brew upgrade %s\n", sitiBrewFormula)
		if err := runCmd("brew", "upgrade", sitiBrewFormula); err != nil {
			return false, fmt.Errorf("升级 siti-cli 失败: %w", err)
		}
		fmt.Println("✓ done")
		return true, nil
	case "standalone":
		dir := os.ExpandEnv("$HOME/.siti-cli")
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return false, fmt.Errorf("未找到安装目录: %s", dir)
		}
		fmt.Println("→ git pull")
		c := exec.Command("git", "pull", "--rebase", "--autostash", "origin", "main")
		c.Dir = dir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return false, fmt.Errorf("git pull 失败: %w", err)
		}
		fmt.Println("✓ done")
		return true, nil
	default:
		fmt.Printf("! 未知安装方式: %s\n", installMethod)
		fmt.Printf("  Homebrew: brew upgrade %s\n", sitiBrewFormula)
		fmt.Println("  独立安装: cd ~/.siti-cli && git pull")
		return false, nil
	}
}

// brewFormulaOutdated distinguishes Homebrew's three relevant outcomes:
// no output means current, package output means outdated (exit 1 is normal),
// and exit 1 without package output is an actual error such as an untrusted tap.
func brewFormulaOutdated(formula string) (bool, error) {
	c := exec.Command("brew", "outdated", "--formula", formula)
	out, err := c.Output()
	if len(bytes.TrimSpace(out)) > 0 {
		return true, nil
	}
	if err == nil {
		return false, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if detail := strings.TrimSpace(string(exitErr.Stderr)); detail != "" {
			return false, errors.New(detail)
		}
	}
	return false, err
}
