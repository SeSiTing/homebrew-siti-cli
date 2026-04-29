package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// sectionSelf upgrades siti-cli. Returns true if an update was applied.
func sectionSelf(cmd *cobra.Command) bool {
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
		outdated, _ := exec.Command("brew", "outdated", "siti-cli").Output()
		if len(outdated) == 0 {
			fmt.Println("✓ 已是最新版本")
			return false
		}
		fmt.Println("→ brew upgrade siti-cli")
		if _, err := exec.LookPath("brew"); err == nil {
			runCmd("brew", "update")
		}
		runCmd("brew", "upgrade", "siti-cli")
		fmt.Println("✓ done")
		return true
	case "standalone":
		dir := os.ExpandEnv("$HOME/.siti-cli")
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "✗ 未找到安装目录: %s\n", dir)
			return false
		}
		fmt.Println("→ git pull")
		c := exec.Command("git", "pull", "--rebase", "--autostash", "origin", "main")
		c.Dir = dir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ git pull 失败: %v\n", err)
			return false
		}
		fmt.Println("✓ done")
		return true
	default:
		fmt.Printf("! 未知安装方式: %s\n", installMethod)
		fmt.Println("  Homebrew: brew upgrade siti-cli")
		fmt.Println("  独立安装: cd ~/.siti-cli && git pull")
		return false
	}
}
