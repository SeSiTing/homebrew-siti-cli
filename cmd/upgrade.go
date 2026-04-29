package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	upgradeSelf     bool
	upgradeBrew     bool
	upgradeNpm      bool
	upgradeGem      bool
	upgradeAll      bool
	upgradeDryRun   bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "升级 siti-cli 或系统包管理器中的包",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		hasTarget := upgradeSelf || upgradeBrew || upgradeNpm || upgradeGem || upgradeAll

		runSelf := upgradeSelf || upgradeAll
		runBrew := upgradeBrew || upgradeAll || !hasTarget
		runNpm  := upgradeNpm  || upgradeAll || !hasTarget
		runGem  := upgradeGem  || upgradeAll || !hasTarget

		t0 := time.Now()
		var sections []string

		fmt.Println()

		if runSelf {
			sections = append(sections, "self")
			if err := sectionSelf(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "✗ siti-cli 升级失败: %v\n", err)
			}
			fmt.Println()
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

		if runGem {
			sections = append(sections, "gem")
			if err := sectionGem(); err != nil {
				fmt.Fprintf(os.Stderr, "✗ gem: %v\n", err)
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
	upgradeCmd.Flags().BoolVar(&upgradeGem, "gem", false, "仅升级 Ruby gem")
	upgradeCmd.Flags().BoolVar(&upgradeAll, "all", false, "升级所有包管理器（含 self）")
	upgradeCmd.Flags().BoolVarP(&upgradeDryRun, "dry-run", "n", false, "仅预览，不执行更新")
	rootCmd.AddCommand(upgradeCmd)
}
