package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var brewupInteractive bool

var brewCmd = &cobra.Command{
	Use:   "brew",
	Short: "Homebrew 辅助命令",
}

var brewUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Homebrew 一键升级全流程（update/upgrade/cleanup）",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("────────────────────────────────────────")
		fmt.Println("Homebrew 一键升级全流程")
		fmt.Println("────────────────────────────────────────")

		start := time.Now()
		reader := bufio.NewReader(os.Stdin)

		type step struct {
			name string
			fn   func() error
		}

		steps := []step{
			{"更新 Homebrew 自身", func() error { return runCmd("brew", "update") }},
			{"升级所有 formula", func() error { return runCmd("brew", "upgrade") }},
			{"升级所有 cask（包括自动更新的应用）", func() error {
				return runCmd("brew", "upgrade", "--cask", "--greedy")
			}},
			{"删除无用依赖", func() error { return runCmd("brew", "autoremove") }},
			{"清理缓存和旧版本", func() error { return runCmd("brew", "cleanup", "--prune=all") }},
		}

		total := len(steps)
		var errs []string

		for i, s := range steps {
			fmt.Printf("\n→ [%d/%d] %s\n", i+1, total, s.name)

			if brewupInteractive {
				fmt.Print("? 执行此步骤? [Y/n] ")
				line, _ := reader.ReadString('\n')
				line = strings.TrimSpace(strings.ToLower(line))
				if line == "n" || line == "no" {
					fmt.Printf("↷ skip: %s\n", s.name)
					continue
				}
			}

			if err := s.fn(); err != nil {
				msg := fmt.Sprintf("步骤 %d 失败: %v", i+1, err)
				errs = append(errs, msg)
				fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
			} else {
				fmt.Printf("✓ [%d/%d] 完成\n", i+1, total)
			}
		}

		elapsed := time.Since(start).Round(time.Second)
		fmt.Printf("\n────────────────────────────────────────\n")
		fmt.Printf("took %s\n", elapsed)

		if len(errs) > 0 {
			fmt.Printf("\n! 执行过程中遇到 %d 个错误:\n", len(errs))
			for _, e := range errs {
				fmt.Println("  •", e)
			}
			return fmt.Errorf("升级流程完成，但存在错误")
		}

		fmt.Println("\n✓ 所有步骤执行成功")
		return nil
	},
}

func init() {
	brewUpCmd.Flags().BoolVarP(&brewupInteractive, "interactive", "i", false, "逐步确认每个步骤")
	brewCmd.AddCommand(brewUpCmd)
	rootCmd.AddCommand(brewCmd)
}
