package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/SeSiTing/siti-cli/internal/config"
	"github.com/SeSiTing/siti-cli/internal/shell"
	"github.com/spf13/cobra"
)

var aiListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用的 AI 服务商",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		providers, err := config.ReadProviders()
		if err != nil {
			return fmt.Errorf("读取配置失败: %w", err)
		}
		skip := config.ReadSkipList()
		current := config.CurrentProvider(providers)
		fmt.Println("可用的 AI 服务商:")
		for _, p := range providers {
			marker := ""
			if p.DisplayName() == current {
				marker = " ← 当前 Claude"
			}
			if p.IsSkipped(skip) {
				fmt.Printf("  ○ %-15s [跳过]\n", p.DisplayName())
				continue
			}
			keyState := "缺少 " + p.AuthTokenVar
			if os.Getenv(p.AuthTokenVar) != "" {
				keyState = "key 已设置"
			}
			fmt.Printf("  • %-15s [%s] %s%s\n", p.DisplayName(), clientSummary(p), keyState, marker)
		}
		return nil
	},
}

var aiCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "显示各客户端当前 AI 配置",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		clients, err := parseAIClients(aiCurrentClient, aiClients{claude: true, grok: true, codex: true})
		if err != nil {
			return err
		}
		providers, _ := config.ReadProviders()
		if clients.claude {
			printShellCurrent("Claude", os.Getenv("ANTHROPIC_BASE_URL"), os.Getenv("ANTHROPIC_MODEL"), providers)
		}
		if clients.grok {
			printGrokCurrent(providers)
		}
		if clients.codex {
			status, err := config.ReadCodexStatus()
			if err != nil {
				return fmt.Errorf("读取 Codex 配置失败: %w", err)
			}
			if status.Managed {
				fmt.Printf("Codex   %-10s %-16s 全局（CLI/Desktop/IDE）\n", status.ProviderName, status.Model)
			} else {
				fmt.Println("Codex   未由 siti 管理")
			}
		}
		return nil
	},
}

var aiDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "检查 AI 切换配置与凭证",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		providers, err := config.ReadProviders()
		if err != nil {
			return fmt.Errorf("读取配置失败: %w", err)
		}
		issues := 0
		fmt.Println("AI 配置检查:")
		for _, p := range providers {
			if p.IsSkipped(config.ReadSkipList()) {
				continue
			}
			if os.Getenv(p.AuthTokenVar) == "" {
				fmt.Printf("  ! %-12s 缺少 %s\n", p.DisplayName(), p.AuthTokenVar)
				issues++
			} else {
				fmt.Printf("  ✓ %-12s shell key 已设置\n", p.DisplayName())
			}
			if p.SupportsCodex() {
				has, credErr := config.HasCredential(p.DisplayName())
				switch {
				case credErr != nil:
					fmt.Printf("  ! %-12s 系统凭证读取失败: %v\n", p.DisplayName(), credErr)
					issues++
				case has:
					fmt.Printf("  ✓ %-12s Codex 系统凭证已保存\n", p.DisplayName())
				default:
					fmt.Printf("  ! %-12s Codex 系统凭证未保存\n", p.DisplayName())
					issues++
				}
			}
		}
		if config.GrokConfigReady() {
			fmt.Println("  ✓ Grok 模型入口已就绪")
		} else {
			fmt.Println("  → Grok 模型入口尚未初始化（首次 switch 时会自动创建）")
		}
		status, err := config.ReadCodexStatus()
		if err != nil {
			fmt.Printf("  ! Codex 配置读取失败: %v\n", err)
			issues++
		} else if status.Managed {
			fmt.Printf("  ✓ Codex 全局配置: %s / %s\n", status.ProviderName, status.Model)
		}
		if issues == 0 {
			fmt.Println("✓ 未发现问题")
		} else {
			fmt.Printf("! 发现 %d 个需要处理的问题\n", issues)
		}
		return nil
	},
}

var aiTestCmd = &cobra.Command{
	Use:   "test",
	Short: "测试当前 Claude API 配置是否可用",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		baseURL := os.Getenv("ANTHROPIC_BASE_URL")
		authToken := os.Getenv("ANTHROPIC_AUTH_TOKEN")
		fmt.Println("→ 测试 Claude API 配置...")
		if baseURL == "" {
			return fmt.Errorf("ANTHROPIC_BASE_URL 未设置\n请运行 'siti ai switch' 选择服务商")
		}
		if authToken == "" {
			return fmt.Errorf("ANTHROPIC_AUTH_TOKEN 未设置\n请运行 'siti ai switch' 选择服务商")
		}
		fmt.Println("  ✓ BASE_URL:", baseURL)
		fmt.Println("  ✓ AUTH_TOKEN: 已设置")
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Head(baseURL)
		if err != nil {
			fmt.Println("  ! 连接测试失败:", err)
			return nil
		}
		resp.Body.Close()
		fmt.Printf("  ✓ 连接正常 (HTTP %d)\n", resp.StatusCode)
		return nil
	},
}

var aiClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "清除客户端切换配置",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		clients, err := parseAIClients(aiClearClient, aiClients{claude: true, grok: true})
		if err != nil {
			return err
		}
		if (clients.claude || clients.grok) && os.Getenv("SITI_WRAPPER") == "" && os.Getenv("SITI_EVAL_FILE") == "" {
			return fmt.Errorf("shell wrapper 未加载，当前 shell 无法清除\n请运行: eval \"$(siti init %s)\"", detectShell())
		}
		if clients.codex {
			path, backup, changed, err := config.ClearCodexConfig()
			if err != nil {
				return fmt.Errorf("清除 Codex 全局配置失败: %w", err)
			}
			if changed {
				printErr("✓ 已恢复 Codex 原有全局配置: %s", path)
				printErr("  备份: %s", backup)
			} else {
				printErr("✓ Codex 没有 siti 管理的全局配置")
			}
		}
		keys := make([]string, 0, len(anthropicModelKeys)+len(grokEnvKeys)+3)
		if clients.claude {
			keys = append(keys, "ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_API_KEY", "ANTHROPIC_BASE_URL")
			keys = append(keys, anthropicModelKeys...)
		}
		if clients.grok {
			keys = append(keys, grokEnvKeys...)
		}
		if len(keys) > 0 {
			Eval(c, shell.Unset(keys...))
			printErr("✓ 已清除当前 shell 的 %s 切换变量", selectedShellClients(clients))
		}
		return nil
	},
}

func printShellCurrent(client, baseURL, model string, providers config.ProviderList) {
	if baseURL == "" {
		fmt.Printf("%-7s 未配置\n", client)
		return
	}
	provider := "custom"
	for _, p := range providers {
		candidate := p.BaseURL()
		if candidate == baseURL {
			provider = p.DisplayName()
			break
		}
	}
	if model == "" {
		model = "默认模型"
	}
	fmt.Printf("%-7s %-10s %-16s 当前 shell\n", client, provider, model)
}

func printGrokCurrent(providers config.ProviderList) {
	modelID := os.Getenv("SITI_GROK_MODEL_ID")
	if modelID == "" {
		fmt.Println("Grok    未由 siti 在当前 shell 管理")
		return
	}
	provider := "custom"
	model := modelID
	for _, p := range providers {
		if config.GrokModelID(p) == modelID {
			provider = p.DisplayName()
			model = p.GrokModel()
			break
		}
	}
	fmt.Printf("Grok    %-10s %-16s 当前 shell\n", provider, model)
}
