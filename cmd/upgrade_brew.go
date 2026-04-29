package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
)

func sectionBrewDryRun() {
	fmt.Println("── brew ── (dry-run)")
	if _, err := exec.LookPath("brew"); err != nil {
		fmt.Println("✗ brew 未安装")
		return
	}
	sr, err := scanOutdatedSilent()
	if err != nil {
		fmt.Printf("! 扫描失败: %v\n", err)
		return
	}
	if sr.isEmpty() {
		fmt.Println("✓ 所有 package 已是最新")
		return
	}
	if len(sr.formula) > 0 {
		fmt.Printf("  %d formula 可更新:\n", len(sr.formula))
		for _, p := range sr.formula {
			fmt.Printf("    %s  %s → %s\n", p.name, p.oldVer, p.newVer)
		}
	}
	if len(sr.cask) > 0 {
		fmt.Printf("  %d cask 可更新:\n", len(sr.cask))
		for _, p := range sr.cask {
			fmt.Printf("    %s  %s → %s\n", p.name, p.oldVer, p.newVer)
		}
	}
}

func sectionBrew() error {
	fmt.Println("── brew ──")
	if _, err := exec.LookPath("brew"); err != nil {
		fmt.Println("✗ brew 未安装")
		return nil
	}

	fmt.Println("→ brew update")
	exec.Command("brew", "update").Run()

	before, err := scanOutdatedSilent()
	if err != nil {
		return fmt.Errorf("扫描失败: %w", err)
	}
	if before.isEmpty() {
		fmt.Println("✓ 所有 package 已是最新")
		return nil
	}

	fmt.Printf("! %s 可更新\n", before.summary())
	if len(before.formula) > 0 {
		fmt.Printf("  %d formula:\n", len(before.formula))
		for _, p := range before.formula {
			fmt.Printf("    %s  %s → %s\n", p.name, p.oldVer, p.newVer)
		}
	}
	if len(before.cask) > 0 {
		fmt.Printf("  %d cask:\n", len(before.cask))
		for _, p := range before.cask {
			fmt.Printf("    %s  %s → %s\n", p.name, p.oldVer, p.newVer)
		}
	}

	fmt.Println("\n→ brew upgrade")
	exec.Command("brew", "upgrade").Run()

	if len(before.cask) > 0 {
		fmt.Println("→ brew upgrade --cask --greedy")
		exec.Command("brew", "upgrade", "--cask", "--greedy").Run()
	}

	fmt.Println("→ brew autoremove + cleanup")
	exec.Command("brew", "autoremove").Run()
	var buf bytes.Buffer
	runCmdTee(&buf, "brew", "cleanup", "--prune=all")
	mb := parseCleanupMB(buf.String())
	if mb != "" {
		fmt.Printf("  清理: %s\n", mb)
	}

	after, _ := scanOutdatedSilent()
	upgraded := diffScan(before, after)
	if !upgraded.isEmpty() {
		fmt.Printf("✓ 已更新 %s\n", upgraded.summary())
	} else {
		fmt.Println("✓ 已更新")
	}
	return nil
}
