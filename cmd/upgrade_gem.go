package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type gemPkg struct {
	name    string
	current string
	latest  string
}

func scanGemOutdated() ([]gemPkg, error) {
	if _, err := exec.LookPath("gem"); err != nil {
		return nil, nil
	}
	out, err := exec.Command("gem", "outdated").Output()
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`^(\S+)\s+\((\S+)\s+<\s+(\S+)\)`)
	var pkgs []gemPkg
	for _, line := range strings.Split(string(out), "\n") {
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		pkgs = append(pkgs, gemPkg{name: m[1], current: m[2], latest: m[3]})
	}
	return pkgs, nil
}

func sectionGem() error {
	fmt.Println("── gem ──")
	if _, err := exec.LookPath("gem"); err != nil {
		fmt.Println("✗ gem 未安装")
		return nil
	}

	pkgs, err := scanGemOutdated()
	if err != nil {
		return fmt.Errorf("扫描失败: %w", err)
	}
	if len(pkgs) == 0 {
		fmt.Println("✓ 所有 gem 已是最新")
		return nil
	}

	fmt.Printf("! %d 个 gem 可更新:\n", len(pkgs))
	for _, p := range pkgs {
		fmt.Printf("    %s  %s → %s\n", p.name, p.current, p.latest)
	}

	if upgradeDryRun {
		fmt.Println("\n(dry-run, 跳过更新)")
		return nil
	}

	fmt.Println("\n→ gem update")
	c := exec.Command("gem", "update")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("gem update 失败: %w", err)
	}
	fmt.Printf("✓ 已更新 %d 个 gem\n", len(pkgs))
	return nil
}
