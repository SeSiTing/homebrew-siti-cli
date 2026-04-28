package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SeSiTing/siti-cli/internal/shell"
	"github.com/spf13/cobra"
)

var initAuto bool

var initCmd = &cobra.Command{
	Use:       "init [zsh|bash|fish]",
	Short:     "输出 shell wrapper 配置（添加到 shell 配置文件）",
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"zsh", "bash", "sh", "fish"},
	Long: `输出 shell wrapper 函数，使 'siti ai switch' 和 'siti proxy on/off'
能够修改当前终端的环境变量。

用法:
  # 查看内容
  siti init zsh

  # 当前会话生效
  eval "$(siti init zsh)"

  # 自动追加到配置文件并 source（推荐）
  siti init zsh --auto`,
	RunE: func(cmd *cobra.Command, args []string) error {
		shellType := "zsh"
		if len(args) > 0 {
			shellType = args[0]
		}

		switch shellType {
		case "zsh", "bash", "sh":
		case "fish":
		default:
			return fmt.Errorf("不支持的 shell 类型: %s\n支持: zsh, bash, fish", shellType)
		}

		if initAuto {
			return autoInstall(shellType)
		}

		fmt.Println(shell.WrapperFor(shellType))
		return nil
	},
}

// autoInstall appends the wrapper to the shell config file if not already present.
func autoInstall(shellType string) error {
	rc := shellRCPath(shellType)
	if rc == "" {
		return fmt.Errorf("无法找到 shell 配置文件")
	}

	// Normalize sh -> bash for config file purposes
	if shellType == "sh" {
		shellType = "bash"
	}

	data, _ := os.ReadFile(rc)
	content := string(data)

	// Check if wrapper is already present
	if strings.Contains(content, "siti shell wrapper") || strings.Contains(content, "siti init") {
		fmt.Fprintf(os.Stderr, "✓ %s 已包含 siti wrapper，无需重复添加\n", rc)
		return nil
	}

	// Append wrapper init line
	initLine := fmt.Sprintf(`eval "$(siti init %s)"`, shellType)
	if shellType == "fish" {
		initLine = "siti init fish | source"
	}

	entry := fmt.Sprintf("\n# siti shell wrapper\n%s\n", initLine)

	f, err := os.OpenFile(rc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("写入 %s 失败: %w", rc, err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("写入 %s 失败: %w", rc, err)
	}

	fmt.Fprintf(os.Stderr, "✓ 已追加到 %s\n", rc)
	fmt.Fprintf(os.Stderr, "→ 当前会话生效: source %s\n", filepath.Base(rc))
	return nil
}

func init() {
	initCmd.Flags().BoolVar(&initAuto, "auto", false, "自动追加到 shell 配置文件")
	rootCmd.AddCommand(initCmd)
}
