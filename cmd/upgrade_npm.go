package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type npmPkg struct {
	name    string
	current string
	latest  string
}

func scanNpmOutdated() ([]npmPkg, error) {
	if _, err := exec.LookPath("npm"); err != nil {
		return nil, nil
	}
	out, err := exec.Command("npm", "outdated", "-g", "--parseable").Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// npm outdated returns exit 1 when packages are outdated; data is in stdout (already in out)
			if len(out) == 0 {
				out = exitErr.Stderr
			}
		} else {
			return nil, err
		}
	}
	var pkgs []npmPkg
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		// Format: path/to/pkg:pkg@latest:pkg@current:pkg@wanted:location
		parts := strings.SplitN(line, ":", 5)
		if len(parts) < 4 {
			continue
		}

		// extractPkgVer splits "pkg@ver" into (pkg, ver), handling @scope/name@ver
		extractPkgVer := func(s string) (string, string) {
			start := 0
			if strings.HasPrefix(s, "@") {
				if i := strings.Index(s[1:], "@"); i >= 0 {
					start = i + 1
				}
			}
			if i := strings.Index(s[start:], "@"); i >= 0 {
				return s[:start+i], s[start+i+1:]
			}
			// For scoped packages, start already points past the name@ boundary
			if start > 0 {
				return s[:start], s[start:]
			}
			return s, ""
		}

		name, latest := extractPkgVer(parts[1])
		_, current := extractPkgVer(parts[2])
		pkgs = append(pkgs, npmPkg{name: name, current: current, latest: latest})
	}
	return pkgs, nil
}

func sectionNpm() error {
	fmt.Println("── npm global ──")
	if _, err := exec.LookPath("npm"); err != nil {
		fmt.Println("✗ npm 未安装")
		return nil
	}

	pkgs, err := scanNpmOutdated()
	if err != nil {
		return fmt.Errorf("扫描失败: %w", err)
	}
	if len(pkgs) == 0 {
		fmt.Println("✓ 所有 npm global 包已是最新")
		return nil
	}

	fmt.Printf("! %d 个包可更新:\n", len(pkgs))
	for _, p := range pkgs {
		fmt.Printf("    %s  %s → %s\n", p.name, p.current, p.latest)
	}

	if upgradeDryRun {
		fmt.Println("\n(dry-run, 跳过更新)")
		return nil
	}

	fmt.Println("\n→ npm update -g")
	var names []string
	for _, p := range pkgs {
		names = append(names, p.name)
	}
	args := append([]string{"update", "-g"}, names...)
	c := exec.Command("npm", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("npm update 失败: %w", err)
	}
	fmt.Printf("✓ 已更新 %d 个 npm global 包\n", len(pkgs))
	return nil
}
