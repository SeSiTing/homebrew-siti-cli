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

// parseOutdated parses "brew outdated --verbose [--cask] [--greedy]" output.
// Format: "name (oldVer) < newVer" or "name (oldVer) != newVer" (optional size in parens).
func parseOutdated(out string) []pkgInfo {
	var pkgs []pkgInfo
	re := regexp.MustCompile(`^(\S+)\s+\(([^)]+)\)\s+(?:<|!=)\s+(\S+)(?:\s+\(([^)]+)\))?`)
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
		if err := runBrewUpdate(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ brew update 失败: %v\n", err)
			return err
		}
		fmt.Println("✓ 完成")

		// Step 2: scan outdated (pre-upgrade)
		fmt.Println("\n→ 扫描待更新 package")
		before, err := scanOutdated()
		if err != nil {
			fmt.Fprintf(os.Stderr, "! 扫描失败: %v\n", err)
			fmt.Println("→ 继续执行升级...")
		}

		hasUpdates := !before.isEmpty()

		if !hasUpdates {
			fmt.Println("✓ 所有 package 已是最新版本")
		} else {
			// Show preview
			fmt.Println()
			if len(before.formula) > 0 {
				fmt.Printf("  %d formula:\n", len(before.formula))
				for _, p := range before.formula {
					fmt.Println(outdatedLine(p))
				}
			}
			if len(before.cask) > 0 {
				fmt.Printf("  %d cask:\n", len(before.cask))
				for _, p := range before.cask {
					fmt.Println(outdatedLine(p))
				}
			}
			fmt.Printf("\n  将更新 %s\n", before.summary())

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
			{"升级所有 formula", !hasUpdates || len(before.formula) == 0, func() error {
				return runCmd("brew", "upgrade")
			}},
			{"升级所有 cask（--greedy）", !hasUpdates || len(before.cask) == 0, func() error {
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

		// Post-upgrade scan: determine what was actually upgraded
		fmt.Println("\n→ 验证更新结果")
		after, err := scanOutdatedSilent()
		if err != nil {
			fmt.Println("↷ 验证扫描失败，使用升级前数据汇总")
			after = scanResult{} // assume all upgraded
		}
		upgraded := diffScan(before, after)
		if !upgraded.isEmpty() {
			fmt.Println("✓ 更新完成:")
			if len(upgraded.formula) > 0 {
				fmt.Printf("  %d formula:\n", len(upgraded.formula))
				for _, p := range upgraded.formula {
					fmt.Printf("    %s  %s → %s\n", p.name, p.oldVer, p.newVer)
				}
			}
			if len(upgraded.cask) > 0 {
				fmt.Printf("  %d cask:\n", len(upgraded.cask))
				for _, p := range upgraded.cask {
					fmt.Printf("    %s  %s → %s\n", p.name, p.oldVer, p.newVer)
				}
			}
		} else if hasUpdates {
			fmt.Println("✓ 所有 package 已更新")
		} else {
			fmt.Println("↷ 无需验证（无待更新项）")
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

		if !upgraded.isEmpty() {
			parts := make([]string, 0, 2)
			if len(upgraded.formula) > 0 {
				parts = append(parts, fmt.Sprintf("%d formula", len(upgraded.formula)))
			}
			if len(upgraded.cask) > 0 {
				parts = append(parts, fmt.Sprintf("%d cask", len(upgraded.cask)))
			}
			fmt.Printf("已更新: %s\n", strings.Join(parts, " + "))
		} else if hasUpdates {
			fmt.Println("! 无 package 被成功更新（可能已跳过或失败）")
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
	if err := runCmdTee(&fBuf, "brew", "outdated", "--verbose"); err != nil {
		return scan, fmt.Errorf("brew outdated: %w", err)
	}
	scan.formula = parseOutdated(fBuf.String())

	// Scan cask
	var cBuf bytes.Buffer
	if err := runCmdTee(&cBuf, "brew", "outdated", "--cask", "--greedy", "--verbose"); err != nil {
		return scan, fmt.Errorf("brew outdated --cask: %w", err)
	}
	scan.cask = parseOutdated(cBuf.String())

	return scan, nil
}

// scanOutdatedSilent is like scanOutdated but suppresses stdout (no raw brew output).
func scanOutdatedSilent() (scanResult, error) {
	var scan scanResult

	var fBuf bytes.Buffer
	if err := runCmdOutput(&fBuf, "brew", "outdated", "--verbose"); err != nil {
		return scan, fmt.Errorf("brew outdated: %w", err)
	}
	scan.formula = parseOutdated(fBuf.String())

	var cBuf bytes.Buffer
	if err := runCmdOutput(&cBuf, "brew", "outdated", "--cask", "--greedy", "--verbose"); err != nil {
		return scan, fmt.Errorf("brew outdated --cask: %w", err)
	}
	scan.cask = parseOutdated(cBuf.String())

	return scan, nil
}

// diffScan returns packages present in before but absent in after (i.e. successfully upgraded).
func diffScan(before, after scanResult) scanResult {
	afterSet := make(map[string]bool)
	for _, p := range after.formula {
		afterSet[p.name] = true
	}
	for _, p := range after.cask {
		afterSet[p.name] = true
	}
	var diff scanResult
	for _, p := range before.formula {
		if !afterSet[p.name] {
			diff.formula = append(diff.formula, p)
		}
	}
	for _, p := range before.cask {
		if !afterSet[p.name] {
			diff.cask = append(diff.cask, p)
		}
	}
	return diff
}

func init() {
	brewUpCmd.Flags().BoolVarP(&brewupInteractive, "interactive", "i", false, "逐步确认每个步骤")
	brewUpCmd.Flags().BoolVarP(&brewupDryRun, "dry-run", "n", false, "仅预览，不执行更新")
	brewCmd.AddCommand(brewUpCmd)
	rootCmd.AddCommand(brewCmd)
}
