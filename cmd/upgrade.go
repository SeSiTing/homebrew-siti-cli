package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	upgradeSelf   bool
	upgradeBrew   bool
	upgradeNpm    bool
	upgradeBlOps  bool
	upgradeAll    bool
	upgradeDryRun bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "升级 siti-cli 或系统包管理器中的包",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		hasTarget := upgradeSelf || upgradeBrew || upgradeNpm || upgradeBlOps || upgradeAll

		// Default (no flags): self first, then brew + npm.
		runSelf := upgradeSelf || upgradeAll || !hasTarget
		runBrew := upgradeBrew || upgradeAll || !hasTarget
		runNpm := upgradeNpm || upgradeAll || !hasTarget
		runBlOps := upgradeBlOps || upgradeAll

		t0 := time.Now()
		var sections []string

		fmt.Println()

		if runSelf {
			sections = append(sections, "self")
			updated, err := sectionSelf(cmd)
			if err != nil {
				return err
			}
			fmt.Println()
			if updated {
				fmt.Println("→ siti-cli 已更新，建议重新运行: siti upgrade")
				fmt.Printf("→ 完成 (took %s) [self]\n", time.Since(t0).Round(time.Second))
				return nil
			}
		}

		if runBrew {
			sections = append(sections, "brew")
			if upgradeDryRun {
				sectionBrewDryRun()
			} else {
				if err := sectionBrew(); err != nil {
					fmt.Fprintf(os.Stderr, "✗ brew: %v\n", err)
				}
			}
			fmt.Println()
		}

		if runNpm {
			sections = append(sections, "npm")
			if err := sectionNpm(); err != nil {
				fmt.Fprintf(os.Stderr, "✗ npm: %v\n", err)
			}
			fmt.Println()
		}

		if runBlOps {
			sections = append(sections, "bl-ops")
			if err := sectionBlOps(); err != nil {
				fmt.Fprintf(os.Stderr, "✗ bl-ops: %v\n", err)
			}
			fmt.Println()
		}

		elapsed := time.Since(t0).Round(time.Second)
		fmt.Printf("→ 完成 (took %s) [%s]\n", elapsed, strings.Join(sections, " + "))
		return nil
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeSelf, "self", false, "仅升级 siti-cli 自身")
	upgradeCmd.Flags().BoolVar(&upgradeBrew, "brew", false, "仅升级 Homebrew 包")
	upgradeCmd.Flags().BoolVar(&upgradeNpm, "npm", false, "仅升级 npm 全局包")
	upgradeCmd.Flags().BoolVar(&upgradeBlOps, "bl-ops", false, "升级 bl-ops 工具")
	upgradeCmd.Flags().BoolVar(&upgradeAll, "all", false, "升级所有目标（含 self、brew、npm、bl-ops）")
	upgradeCmd.Flags().BoolVarP(&upgradeDryRun, "dry-run", "n", false, "仅预览，不执行更新")
	rootCmd.AddCommand(upgradeCmd)
}
