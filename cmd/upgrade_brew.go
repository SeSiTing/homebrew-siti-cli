package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
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
	// Filter out siti-cli (self handles it separately).
	var formula []pkgInfo
	for _, p := range sr.formula {
		if !strings.HasSuffix(p.name, "siti-cli") {
			formula = append(formula, p)
		}
	}
	if len(formula) == 0 && len(sr.cask) == 0 {
		fmt.Println("✓ 所有 package 已是最新")
		return
	}
	if len(formula) > 0 {
		fmt.Printf("  %d formula 可更新:\n", len(formula))
		for _, p := range formula {
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

	before, err := scanOutdatedSilent()
	if err != nil {
		return fmt.Errorf("扫描失败: %w", err)
	}
	if before.isEmpty() {
		fmt.Println("✓ 所有 package 已是最新")
		return nil
	}

	// Filter out siti-cli from formula (self handles it separately).
	var formula []pkgInfo
	for _, p := range before.formula {
		if !strings.HasSuffix(p.name, "siti-cli") {
			formula = append(formula, p)
		}
	}

	if len(formula) == 0 && len(before.cask) == 0 {
		fmt.Println("✓ 所有 package 已是最新")
		return nil
	}

	parts := make([]string, 0, 2)
	if len(formula) > 0 {
		parts = append(parts, fmt.Sprintf("%d formula", len(formula)))
	}
	if len(before.cask) > 0 {
		parts = append(parts, fmt.Sprintf("%d cask", len(before.cask)))
	}
	fmt.Printf("! %s 可更新\n", strings.Join(parts, " + "))
	if len(formula) > 0 {
		fmt.Printf("  %d formula:\n", len(formula))
		for _, p := range formula {
			fmt.Printf("    %s  %s → %s\n", p.name, p.oldVer, p.newVer)
		}
	}
	if len(before.cask) > 0 {
		fmt.Printf("  %d cask:\n", len(before.cask))
		for _, p := range before.cask {
			fmt.Printf("    %s  %s → %s\n", p.name, p.oldVer, p.newVer)
		}
	}

	if len(formula) > 0 {
		fmt.Println("\n→ brew upgrade")
		var names []string
		for _, p := range formula {
			names = append(names, p.name)
		}
		args := append([]string{"upgrade"}, names...)
		runCmd("brew", args...)
	}

	if len(before.cask) > 0 {
		fmt.Println("→ brew upgrade --cask --greedy")
		runCmd("brew", "upgrade", "--cask", "--greedy")
	}

	fmt.Println("→ brew autoremove + cleanup")
	runCmd("brew", "autoremove")
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
