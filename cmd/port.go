package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	devPorts  = []int{3000, 5173, 8080, 8000, 4200, 4000, 3001, 5000, 9000}
	dbPorts   = []int{3306, 5432, 27017, 6379, 5984, 9200, 9042, 7000}
	webPorts  = []int{80, 443, 8000, 8080, 8888, 9000, 3000}
	javaPorts = rangePorts(8080, 8090)

	defaultPorts = append(append(append(
		rangePorts(2024, 2030),
		rangePorts(8000, 8010)...),
		rangePorts(8080, 8090)...),
		rangePorts(9000, 9010)...)
)

var (
	kpCheckOnly bool
	kpDev       bool
	kpDB        bool
	kpWeb       bool
	kpJava      bool
	kpAll       bool
)

var portCmd = &cobra.Command{
	Use:   "port",
	Short: "端口管理",
}

var portKillCmd = &cobra.Command{
	Use:   "kill [ports...]",
	Short: "释放被占用的端口",
	Long: `扫描并释放指定端口的占用进程。

端口格式:
  单个:    siti port kill 8080
  多个:    siti port kill 8080 9000
  范围:    siti port kill 3000-3010
  逗号:    siti port kill 3000,5000,8080

预设组:
  --dev    开发端口 (3000/5173/8080 等)
  --db     数据库端口 (3306/5432/27017 等)
  --web    Web 端口 (80/443/8000/8080 等)
  --java   Java 端口 (8080-8090)
  --all    显示所有占用端口，逐个确认`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ports, err := resolvePorts(args)
		if err != nil {
			return err
		}

		if kpCheckOnly {
			fmt.Println("→ 扫描端口占用（检查模式）...")
		} else {
			fmt.Println("→ 扫描端口占用...")
		}

		type entry struct {
			port    int
			pids    []string
			cmdLine string
			label   string
		}

		var occupied []entry
		for _, port := range ports {
			pids := pidsOnPort(port)
			if len(pids) == 0 {
				continue
			}
			cmdLine := processName(pids[0])
			occupied = append(occupied, entry{
				port:    port,
				pids:    pids,
				cmdLine: cmdLine,
				label:   processLabel(cmdLine),
			})
		}

		if len(occupied) == 0 {
			fmt.Printf("\n✓ 扫描了 %d 个端口，没有发现占用\n", len(ports))
			return nil
		}

		fmt.Printf("\n────────────────────────────────────────\n")
		fmt.Printf("! 发现 %d 个端口被占用:\n\n", len(occupied))
		for _, e := range occupied {
			fmt.Printf("  %d  %s  PID %s\n", e.port, e.label, strings.Join(e.pids, ","))
			fmt.Printf("       %s\n\n", truncate(e.cmdLine, 60))
		}
		fmt.Println("────────────────────────────────────────")

		if kpCheckOnly {
			fmt.Printf("• 检查模式: 扫描了 %d 个端口，未终止任何进程\n", len(ports))
			return nil
		}

		killed := 0
		reader := bufio.NewReader(os.Stdin)

		if kpAll {
			// Per-port confirmation
			for _, e := range occupied {
				fmt.Printf("%d  %s  %s\n", e.port, e.label, truncate(e.cmdLine, 50))
				fmt.Print("  ? 是否清理? [y/N] ")
				line, _ := reader.ReadString('\n')
				if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "y") {
					killed += killPIDs(e.pids)
					fmt.Println("  ✓ 已清理")
				} else {
					fmt.Println("  ↷ skip")
				}
				fmt.Println()
			}
		} else {
			// Batch confirmation
			fmt.Print("? 是否清理以上所有端口? [y/N] ")
			line, _ := reader.ReadString('\n')
			if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "y") {
				fmt.Println("✗ 已取消")
				return nil
			}
			fmt.Println()
			for _, e := range occupied {
				killed += killPIDs(e.pids)
				fmt.Printf("  ✓ 端口 %d 已清理\n", e.port)
			}
		}

		fmt.Printf("\n────────────────────────────────────────\n")
		fmt.Printf("✓ 完成: 扫描了 %d 个端口，终止了 %d 个进程\n", len(ports), killed)
		return nil
	},
}

func init() {
	portKillCmd.Flags().BoolVar(&kpCheckOnly, "check", false, "仅检查，不终止进程")
	portKillCmd.Flags().BoolVar(&kpDev, "dev", false, "开发端口预设")
	portKillCmd.Flags().BoolVar(&kpDB, "db", false, "数据库端口预设")
	portKillCmd.Flags().BoolVar(&kpWeb, "web", false, "Web 端口预设")
	portKillCmd.Flags().BoolVar(&kpJava, "java", false, "Java 端口预设 (8080-8090)")
	portKillCmd.Flags().BoolVar(&kpAll, "all", false, "扫描所有占用端口并逐个确认")
	portCmd.AddCommand(portKillCmd)
	rootCmd.AddCommand(portCmd)
}

// resolvePorts returns the port list based on flags and arguments.
func resolvePorts(args []string) ([]int, error) {
	switch {
	case kpDev:
		fmt.Println("→ 使用开发端口预设")
		return devPorts, nil
	case kpDB:
		fmt.Println("→ 使用数据库端口预设")
		return dbPorts, nil
	case kpWeb:
		fmt.Println("→ 使用 Web 端口预设")
		return webPorts, nil
	case kpJava:
		fmt.Println("→ 使用 Java 端口预设 (8080-8090)")
		return javaPorts, nil
	case kpAll:
		return allListeningPorts(), nil
	case len(args) == 0:
		return defaultPorts, nil
	default:
		return parsePortArgs(args)
	}
}

func parsePortArgs(args []string) ([]int, error) {
	var ports []int
	for _, arg := range args {
		// comma-separated
		for _, part := range strings.Split(arg, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			// range: 3000-3010
			if idx := strings.Index(part, "-"); idx > 0 {
				start, err1 := strconv.Atoi(part[:idx])
				end, err2 := strconv.Atoi(part[idx+1:])
				if err1 != nil || err2 != nil || start > end {
					return nil, fmt.Errorf("无效的端口范围: %s", part)
				}
				ports = append(ports, rangePorts(start, end)...)
			} else {
				p, err := strconv.Atoi(part)
				if err != nil {
					return nil, fmt.Errorf("无效的端口: %s", part)
				}
				ports = append(ports, p)
			}
		}
	}
	return ports, nil
}

func rangePorts(from, to int) []int {
	ports := make([]int, 0, to-from+1)
	for i := from; i <= to; i++ {
		ports = append(ports, i)
	}
	return ports
}

func allListeningPorts() []int {
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P").Output()
	if err != nil {
		return nil
	}
	seen := map[int]bool{}
	var ports []int
	for _, line := range strings.Split(string(out), "\n")[1:] {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		addr := fields[8]
		idx := strings.LastIndex(addr, ":")
		if idx < 0 {
			continue
		}
		p, err := strconv.Atoi(addr[idx+1:])
		if err == nil && !seen[p] {
			seen[p] = true
			ports = append(ports, p)
		}
	}
	return ports
}

func pidsOnPort(port int) []string {
	out, err := exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%d", port)).Output()
	if err != nil {
		return nil
	}
	var pids []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			pids = append(pids, line)
		}
	}
	return pids
}

func processName(pid string) string {
	out, err := exec.Command("ps", "-p", pid, "-o", "args=").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func processLabel(cmdLine string) string {
	lower := strings.ToLower(cmdLine)
	switch {
	case strings.Contains(lower, "java"):
		return "[java]"
	case strings.Contains(lower, "python"):
		return "[py]"
	case strings.Contains(lower, "node"):
		return "[node]"
	case strings.Contains(lower, "docker"):
		return "[docker]"
	case strings.Contains(lower, "postgres"):
		return "[pg]"
	case strings.Contains(lower, "mysql"):
		return "[mysql]"
	case strings.Contains(lower, "redis"):
		return "[redis]"
	default:
		return "[other]"
	}
}

func killPIDs(pids []string) int {
	killed := 0
	for _, pid := range pids {
		if err := exec.Command("kill", "-9", pid).Run(); err == nil {
			killed++
		}
	}
	return killed
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
