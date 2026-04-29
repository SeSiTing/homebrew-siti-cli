package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func sectionSelf(cmd *cobra.Command) error {
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
		fmt.Println("→ brew upgrade siti-cli")
		if _, err := exec.LookPath("brew"); err == nil {
			exec.Command("brew", "update").Run()
		}
		c := exec.Command("brew", "upgrade", "siti-cli")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
		fmt.Println("✓ done")
		return nil
	case "standalone":
		dir := os.ExpandEnv("$HOME/.siti-cli")
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("未找到安装目录: %s", dir)
		}
		fmt.Println("→ git pull")
		c := exec.Command("git", "pull", "--rebase", "--autostash", "origin", "main")
		c.Dir = dir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
		fmt.Println("✓ done")
		return nil
	default:
		fmt.Printf("! 未知安装方式: %s\n", installMethod)
		fmt.Println("  Homebrew: brew upgrade siti-cli")
		fmt.Println("  独立安装: cd ~/.siti-cli && git pull")
		return nil
	}
}
