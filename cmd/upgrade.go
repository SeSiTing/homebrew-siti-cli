package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "升级 siti-cli 到最新版本",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		installMethod := os.Getenv("INSTALL_METHOD") // set by wrapper if needed

		fmt.Printf("\nsiti-cli 升级工具\n")
		fmt.Printf("   当前版本: v%s\n\n", cmd.Root().Version)

		if installMethod == "" {
			// Auto-detect: if brew knows about siti-cli, use brew.
			if _, err := exec.LookPath("brew"); err == nil {
				out, _ := exec.Command("brew", "list", "--formula", "siti-cli").Output()
				if len(out) > 0 {
					installMethod = "homebrew"
				}
			}
		}

		switch installMethod {
		case "homebrew", "":
			return upgradeViaBrew()
		case "standalone":
			return upgradeViaGit()
		default:
			fmt.Printf("! 安装方式: %s\n", installMethod)
			fmt.Println("请手动更新:")
			fmt.Println("  Homebrew: brew upgrade siti-cli")
			fmt.Println("  独立安装: cd ~/.siti-cli && git pull")
		}
		return nil
	},
}

func upgradeViaBrew() error {
	fmt.Println("→ 通过 Homebrew 更新...")

	start := time.Now()
	if err := runCmd("brew", "update"); err != nil {
		fmt.Fprintf(os.Stderr, "! brew update 失败: %v\n", err)
	}
	if err := runCmd("brew", "upgrade", "siti-cli"); err != nil {
		return fmt.Errorf("升级失败: %w", err)
	}
	fmt.Printf("\n✓ 升级完成 (took %s)\n", time.Since(start).Round(time.Second))
	return nil
}

func upgradeViaGit() error {
	installDir := os.ExpandEnv("$HOME/.siti-cli")
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("未找到安装目录: %s", installDir)
	}

	fmt.Println("→ 通过 Git 更新...")
	if err := runCmdIn(installDir, "git", "pull", "--rebase", "--autostash", "origin", "main"); err != nil {
		return fmt.Errorf("git pull 失败: %w", err)
	}
	fmt.Println("\n✓ 更新完成")
	return nil
}

// runCmd runs a command inheriting stdout/stderr so the user sees live output.
func runCmd(name string, args ...string) error {
	return runCmdIn("", name, args...)
}

// runCmdTee is like runCmd but also tees stdout into buf for post-run parsing.
func runCmdTee(buf *bytes.Buffer, name string, args ...string) error {
	return runCmdInTee("", buf, name, args...)
}

// runCmdIn runs a command in the specified directory (empty = inherit cwd).
// Uses exec.Command.Dir to avoid mutating the process-wide working directory.
func runCmdIn(dir, name string, args ...string) error {
	return runCmdInTee(dir, nil, name, args...)
}

func runCmdInTee(dir string, buf *bytes.Buffer, name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdin = os.Stdin
	c.Dir = dir
	if buf != nil {
		c.Stdout = io.MultiWriter(os.Stdout, buf)
	} else {
		c.Stdout = os.Stdout
	}
	c.Stderr = os.Stderr
	return c.Run()
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
