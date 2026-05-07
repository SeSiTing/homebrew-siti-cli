package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// evalKey is the context key for the per-invocation eval buffer.
type evalKey struct{}

// evalBuffer accumulates shell statements queued by Eval().
type evalBuffer struct{ lines []string }

// Eval queues one or more shell statement strings to be evaluated by the
// parent shell wrapper (exit-code-10 protocol).
//
// Call this from any RunE instead of returning an error-typed directive.
// The command should return nil afterwards; Execute() will detect the buffer
// and exit with code 10 after printing the lines to stdout.
//
// Example:
//
//	cmd.Eval(c, shell.Export("http_proxy", proxyURL))
//	return nil
func Eval(c *cobra.Command, lines ...string) {
	if buf, ok := c.Context().Value(evalKey{}).(*evalBuffer); ok {
		buf.lines = append(buf.lines, lines...)
	}
}

var rootCmd = &cobra.Command{
	Use:   "siti",
	Short: "个人 CLI 工具集",
	Long:  "siti — 个人命令行助手，支持 AI 切换、代理管理、端口清理等便捷操作。",
}

// Execute runs the root command and returns an exit code.
// os.Exit is called exactly once, in main.go.
func Execute(ver string) int {
	rootCmd.Version = ver

	buf := &evalBuffer{}
	ctx := context.WithValue(context.Background(), evalKey{}, buf)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// Real errors are already printed by Cobra.
		return 1
	}

	if len(buf.lines) > 0 {
		if evalFile := os.Getenv("SITI_EVAL_FILE"); evalFile != "" {
			data := []byte{}
			for _, line := range buf.lines {
				data = append(data, line...)
				data = append(data, '\n')
			}
			if err := os.WriteFile(evalFile, data, 0o600); err != nil {
				fmt.Fprintf(os.Stderr, "✗ write eval file: %v\n", err)
				return 1
			}
			return 10
		}

		// No wrapper loaded — show actionable hint.
		printWrapperHint()
		return 0
	}
	return 0
}

// printWrapperHint shows an actionable message when the shell wrapper
// is not loaded (e.g. siti ai switch / proxy on was called directly).
func printWrapperHint() {
	shellType := detectShell()
	rc := shellRCPath(shellType)

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "! shell wrapper 未加载，环境变量不会生效")
	fmt.Fprintln(os.Stderr)
	if shellType == "fish" {
		fmt.Fprintln(os.Stderr, "  当前会话生效:")
		fmt.Fprintln(os.Stderr, "    eval (siti init fish | psub)")
	} else {
		fmt.Fprintln(os.Stderr, "  当前会话生效:")
		fmt.Fprintf(os.Stderr, "    eval \"$(siti init %s)\"\n", shellType)
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  永久生效:")
	fmt.Fprintf(os.Stderr, "    siti init %s --auto\n", shellType)
	if rc != "" {
		fmt.Fprintf(os.Stderr, "    source %s\n", rc)
	}
	fmt.Fprintln(os.Stderr)
}

// detectShell returns the current shell type (zsh, bash, or fish).
func detectShell() string {
	sh := os.Getenv("SHELL")
	base := filepath.Base(sh)
	switch base {
	case "fish":
		return "fish"
	case "bash":
		return "bash"
	default:
		return "zsh"
	}
}

// shellRCPath returns the typical config file path for the given shell.
func shellRCPath(shellType string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch shellType {
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	case "bash":
		// prefer .bashrc over .bash_profile
		if p := filepath.Join(home, ".bashrc"); fileExists(p) {
			return p
		}
		return filepath.Join(home, ".bash_profile")
	default:
		return filepath.Join(home, ".zshrc")
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func init() {
	rootCmd.SilenceErrors = true // handled in Execute()
	rootCmd.SilenceUsage = true  // don't dump usage on every error
	rootCmd.CompletionOptions.DisableDefaultCmd = false
}

// printErr writes a formatted message to stderr (human-visible output).
func printErr(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}
