package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	brewupInteractive bool
	brewupDryRun      bool
)

type pkgInfo struct {
	name   string
	oldVer string
	newVer string
	size   string
}

type scanResult struct {
	formula []pkgInfo
	cask    []pkgInfo
}

func (s *scanResult) isEmpty() bool { return len(s.formula) == 0 && len(s.cask) == 0 }

func (s *scanResult) summary() string {
	parts := make([]string, 0, 2)
	if len(s.formula) > 0 {
		parts = append(parts, fmt.Sprintf("%d formula", len(s.formula)))
	}
	if len(s.cask) > 0 {
		parts = append(parts, fmt.Sprintf("%d cask", len(s.cask)))
	}
	return strings.Join(parts, " + ")
}

// parseOutdated parses "brew outdated [--cask] [--greedy]" output.
// Format: "name oldVer < newVer [(size)]"
func parseOutdated(out string) []pkgInfo {
	var pkgs []pkgInfo
	re := regexp.MustCompile(`^(\S+)\s+(.+)\s+<\s+(\S+)(?:\s+\(([^)]+)\))?`)
	for _, line := range strings.Split(out, "\n") {
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		p := pkgInfo{name: m[1], oldVer: strings.TrimSpace(m[2]), newVer: m[3]}
		if m[4] != "" {
			p.size = m[4]
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// parseCleanupMB extracts "freed approximately X MB" from brew cleanup output.
func parseCleanupMB(buf string) string {
	re := regexp.MustCompile(`freed approximately (\S+) of disk space`)
	if m := re.FindStringSubmatch(buf); m != nil {
		return m[1]
	}
	return ""
}

// outdatedLine formats a pkgInfo as "name oldVer → newVer  (size)"
func outdatedLine(p pkgInfo) string {
	line := fmt.Sprintf("    %s  %s → %s", p.name, p.oldVer, p.newVer)
	if p.size != "" {
		line += "  (" + p.size + ")"
	}
	return line
}

var brewCmd = &cobra.Command{
	Use:   "brew",
	Short: "Homebrew 辅助命令",
}

var brewUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Homebrew 一键升级全流程（update/upgrade/cleanup）",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		header := "Homebrew 一键升级全流程"
		if brewupDryRun {
			header = "Homebrew 一键升级全流程（预览）"
		}
		fmt.Println("────────────────────────────────────────")
		fmt.Println(header)
		fmt.Println("────────────────────────────────────────")

		start := time.Now()

		// Step 1: brew update
		fmt.Println("\n→ 更新 Homebrew 自身")
		if err := runCmd("brew", "update"); err != nil {
			fmt.Fprintf(os.Stderr, "✗ brew update 失败: %v\n", err)
			return err
		}
		fmt.Println("✓ 完成")

		// Step 2: scan outdated
		fmt.Println("\n→ 扫描待更新 package")
		scan, err := scanOutdated()
		if err != nil {
			fmt.Fprintf(os.Stderr, "! 扫描失败: %v\n", err)
			fmt.Println("→ 继续执行升级...")
		}

		hasUpdates := !scan.isEmpty()

		if !hasUpdates {
			fmt.Println("✓ 所有 package 已是最新版本")
		} else {
			// Show preview
			fmt.Println()
			if len(scan.formula) > 0 {
				fmt.Printf("  %d formula:\n", len(scan.formula))
				for _, p := range scan.formula {
					fmt.Println(outdatedLine(p))
				}
			}
			if len(scan.cask) > 0 {
				fmt.Printf("  %d cask:\n", len(scan.cask))
				for _, p := range scan.cask {
					fmt.Println(outdatedLine(p))
				}
			}
			fmt.Printf("\n  将更新 %s\n", scan.summary())

			// Interactive confirmation
			if brewupInteractive {
				fmt.Print("\n  [Enter 继续, Ctrl-C 取消] ")
				bufio.NewReader(os.Stdin).ReadString('\n')
			}
		}

		// dry-run: skip to cleanup summary
		if brewupDryRun {
			fmt.Println("\n────────────────────────────────────────")
			if hasUpdates {
				fmt.Println("! 这是预览，未执行任何更新操作")
				fmt.Println("  运行 'siti brew up' 执行更新")
			} else {
				fmt.Println("✓ 无需更新")
			}
			fmt.Println("────────────────────────────────────────")
			return nil
		}

		// Build dynamic steps
		type stepDef struct {
			label string
			skip  bool
			fn    func() error
		}

		steps := []stepDef{
			{"升级所有 formula", !hasUpdates || len(scan.formula) == 0, func() error {
				return runCmd("brew", "upgrade")
			}},
			{"升级所有 cask（--greedy）", !hasUpdates || len(scan.cask) == 0, func() error {
				return runCmd("brew", "upgrade", "--cask", "--greedy")
			}},
			{"删除无用依赖", !hasUpdates, func() error {
				return runCmd("brew", "autoremove")
			}},
		}

		// Execute steps
		var errs []string
		for _, s := range steps {
			if s.skip {
				continue
			}
			fmt.Printf("\n→ %s\n", s.label)

			if brewupInteractive {
				fmt.Printf("? 执行此步骤? [Y/n] ")
				line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				if strings.TrimSpace(strings.ToLower(line)) == "n" {
					fmt.Printf("↷ skip: %s\n", s.label)
					continue
				}
			}

			if err := s.fn(); err != nil {
				msg := fmt.Sprintf("%s 失败: %v", s.label, err)
				errs = append(errs, msg)
				fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
			} else {
				fmt.Println("✓ 完成")
			}
		}

		// Cleanup
		fmt.Println("\n→ 清理缓存和旧版本")
		var cleanupBuf bytes.Buffer
		if err := runCmdTee(&cleanupBuf, "brew", "cleanup", "--prune=all"); err != nil {
			fmt.Fprintf(os.Stderr, "✗ cleanup 失败: %v\n", err)
		} else {
			fmt.Println("✓ 完成")
		}
		cleanupMB := parseCleanupMB(cleanupBuf.String())

		// Summary
		elapsed := time.Since(start).Round(time.Second)
		fmt.Printf("\n────────────────────────────────────────\n")

		if hasUpdates {
			fTotal := len(scan.formula)
			cTotal := len(scan.cask)
			fmt.Printf("已更新: %d formula + %d cask\n", fTotal, cTotal)
		} else {
			fmt.Println("无需更新")
		}
		if cleanupMB != "" {
			fmt.Printf("清理空间: %s\n", cleanupMB)
		}
		fmt.Printf("总耗时: %s\n", elapsed)
		fmt.Println("────────────────────────────────────────")

		if len(errs) > 0 {
			fmt.Printf("\n! 执行过程中遇到 %d 个错误:\n", len(errs))
			for _, e := range errs {
				fmt.Println("  •", e)
			}
			return fmt.Errorf("升级流程完成，但存在错误")
		}

		fmt.Println("\n✓ 全部完成")
		return nil
	},
}

func scanOutdated() (scanResult, error) {
	var scan scanResult

	// Scan formula
	var fBuf bytes.Buffer
	if err := runCmdTee(&fBuf, "brew", "outdated"); err != nil {
		return scan, fmt.Errorf("brew outdated: %w", err)
	}
	scan.formula = parseOutdated(fBuf.String())

	// Scan cask
	var cBuf bytes.Buffer
	if err := runCmdTee(&cBuf, "brew", "outdated", "--cask", "--greedy"); err != nil {
		return scan, fmt.Errorf("brew outdated --cask: %w", err)
	}
	scan.cask = parseOutdated(cBuf.String())

	return scan, nil
}

func init() {
	brewUpCmd.Flags().BoolVarP(&brewupInteractive, "interactive", "i", false, "逐步确认每个步骤")
	brewUpCmd.Flags().BoolVarP(&brewupDryRun, "dry-run", "n", false, "仅预览，不执行更新")
	brewCmd.AddCommand(brewUpCmd)
	rootCmd.AddCommand(brewCmd)
}
