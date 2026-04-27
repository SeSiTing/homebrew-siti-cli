package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SeSiTing/siti-cli/internal/config"
	"github.com/SeSiTing/siti-cli/internal/shell"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// anthropicModelKeys lists all ANTHROPIC model env vars managed together.
var anthropicModelKeys = []string{
	"ANTHROPIC_MODEL",
	"ANTHROPIC_DEFAULT_SONNET_MODEL",
	"ANTHROPIC_DEFAULT_OPUS_MODEL",
	"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	"ANTHROPIC_REASONING_MODEL",
}

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "管理 AI API 配置切换",
}

// ── siti ai switch ──────────────────────────────────────────────────────────

var aiSwitchCmd = &cobra.Command{
	Use:   "switch [provider]",
	Short: "切换 AI 服务商（仅当前 shell 生效）",
	Long: `切换 AI 服务商，修改当前 shell 的 ANTHROPIC_* 环境变量。
不加参数时启动交互式选择界面。

服务商从 ~/.zshrc 和 ~/.zshenv 中自动发现（定义了 <NAME>_BASE_URL 的变量）。`,
	Args: cobra.MaximumNArgs(1),

	ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		providers, err := config.ReadProviders()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		skip := config.ReadSkipList()
		var completions []cobra.Completion
		for _, p := range providers {
			if !p.IsSkipped(skip) {
				completions = append(completions, cobra.CompletionWithDesc(p.DisplayName(), p.BaseURLVar))
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	},

	RunE: func(c *cobra.Command, args []string) error {
		providers, err := config.ReadProviders()
		if err != nil {
			return fmt.Errorf("读取配置失败: %w", err)
		}
		if len(providers) == 0 {
			return fmt.Errorf("未发现任何 AI 服务商\n提示: 在 ~/.zshrc 或 ~/.zshenv 中定义 <PROVIDER>_BASE_URL 变量")
		}

		skip := config.ReadSkipList()
		var available config.ProviderList
		for _, p := range providers {
			if !p.IsSkipped(skip) {
				available = append(available, p)
			}
		}
		if len(available) == 0 {
			return fmt.Errorf("所有服务商均在跳过列表中 (SITI_AI_SKIP)")
		}

		var providerName string
		if len(args) > 0 {
			providerName = strings.ToUpper(args[0])
		} else {
			providerName, err = selectProvider(available, config.CurrentProvider(available))
			if err != nil || providerName == "" {
				return nil // user cancelled
			}
		}

		p, ok := available.Find(providerName)
		if !ok {
			return fmt.Errorf("服务商 '%s' 不存在，运行 'siti ai list' 查看可用服务商", args[0])
		}

		printErr("✓ 已切换到 %s", p.DisplayName())
		applySwitch(c, p)
		return nil
	},
}

// selectProvider shows an interactive huh selection form.
func selectProvider(providers config.ProviderList, current string) (string, error) {
	options := make([]huh.Option[string], 0, len(providers))
	for _, p := range providers {
		label := p.DisplayName()
		if p.DisplayName() == current {
			label += " ← 当前"
		}
		options = append(options, huh.NewOption(label, p.Name))
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("选择 AI 服务商").
				Options(options...).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return selected, nil
}

// applySwitch queues the eval lines for switching to a provider.
func applySwitch(c *cobra.Command, p config.Provider) {
	lines := []string{
		shell.ExportRef("ANTHROPIC_BASE_URL", p.BaseURLVar),
		shell.ExportRef("ANTHROPIC_AUTH_TOKEN", p.AuthTokenVar),
	}
	if p.ModelVar != "" {
		for _, key := range anthropicModelKeys {
			lines = append(lines, shell.ExportRef(key, p.ModelVar))
		}
	} else {
		lines = append(lines, shell.Unset(anthropicModelKeys...))
	}
	Eval(c, lines...)
}

// ── siti ai list ────────────────────────────────────────────────────────────

var aiListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用的 AI 服务商",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		providers, err := config.ReadProviders()
		if err != nil {
			return fmt.Errorf("读取配置失败: %w", err)
		}
		if len(providers) == 0 {
			fmt.Println("未发现任何 AI 服务商")
			fmt.Println("提示: 在 ~/.zshrc 或 ~/.zshenv 中定义 <PROVIDER>_BASE_URL 变量")
			return nil
		}

		skip := config.ReadSkipList()
		current := config.CurrentProvider(providers)

		fmt.Println("可用的 AI 服务商:")
		for _, p := range providers {
			skipped := p.IsSkipped(skip)
			url := os.Getenv(p.BaseURLVar)
			if url == "" {
				url = "$" + p.BaseURLVar + " (未设置)"
			}
			marker := ""
			if p.DisplayName() == current {
				marker = " ← 当前"
			}
			if skipped {
				fmt.Printf("  ○ %-15s %s [跳过]\n", p.DisplayName(), url)
			} else {
				fmt.Printf("  • %-15s %s%s\n", p.DisplayName(), url, marker)
			}
		}
		return nil
	},
}

// ── siti ai current ─────────────────────────────────────────────────────────

var aiCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "显示当前 AI API 配置",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		baseURL := os.Getenv("ANTHROPIC_BASE_URL")
		authToken := os.Getenv("ANTHROPIC_AUTH_TOKEN")

		fmt.Println("当前 AI API 配置:")

		if baseURL == "" {
			fmt.Println("  ✗ 未配置（ANTHROPIC_BASE_URL 未设置）")
			fmt.Println("  → 运行 'siti ai switch' 选择服务商，或 'source ~/.zshrc' 重新加载")
			return nil
		}

		providers, _ := config.ReadProviders()
		if name := config.CurrentProvider(providers); name != "" {
			fmt.Println("  服务商:", name)
		}
		fmt.Println("  BASE_URL:", baseURL)

		if authToken != "" {
			preview := authToken
			if len(preview) > 20 {
				preview = preview[:20] + "..."
			}
			fmt.Println("  AUTH_TOKEN:", preview)
		}

		for _, key := range anthropicModelKeys {
			if v := os.Getenv(key); v != "" {
				fmt.Printf("  %s: %s\n", key, v)
			}
		}
		return nil
	},
}

// ── siti ai test ─────────────────────────────────────────────────────────────

var aiTestCmd = &cobra.Command{
	Use:   "test",
	Short: "测试当前 AI API 配置是否可用",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		baseURL := os.Getenv("ANTHROPIC_BASE_URL")
		authToken := os.Getenv("ANTHROPIC_AUTH_TOKEN")

		fmt.Println("→ 测试 AI API 配置...")

		if baseURL == "" {
			return fmt.Errorf("ANTHROPIC_BASE_URL 未设置\n请运行 'siti ai switch' 选择服务商")
		}
		if authToken == "" {
			return fmt.Errorf("ANTHROPIC_AUTH_TOKEN 未设置\n请运行 'source ~/.zshrc' 或重新打开终端")
		}

		tokenPreview := authToken[:min(20, len(authToken))]
		fmt.Println("  ✓ BASE_URL:   ", baseURL)
		fmt.Printf("  ✓ AUTH_TOKEN:  %s...\n", tokenPreview)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Head(baseURL)
		if err != nil {
			fmt.Println("  ! 连接测试失败:", err)
		} else {
			resp.Body.Close()
			fmt.Printf("  ✓ 连接正常 (HTTP %d)\n", resp.StatusCode)
		}
		return nil
	},
}

// ── siti ai clear ────────────────────────────────────────────────────────────

var aiClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "清除 ANTHROPIC_* 环境变量（切换到 OAuth 登录模式）",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		printErr("✓ 已清除 ANTHROPIC 环境变量")
		printErr("→ 运行 \"claude login\" 切换到 OAuth 登录模式")

		keys := append(
			[]string{"ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_API_KEY", "ANTHROPIC_BASE_URL"},
			anthropicModelKeys...,
		)
		lines := make([]string, 0, len(keys))
		for _, k := range keys {
			lines = append(lines, shell.Unset(k))
		}
		Eval(c, lines...)
		return nil
	},
}

func init() {
	aiCmd.AddCommand(aiSwitchCmd, aiListCmd, aiCurrentCmd, aiTestCmd, aiClearCmd)
	rootCmd.AddCommand(aiCmd)
}

