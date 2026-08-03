package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/SeSiTing/siti-cli/internal/config"
	"github.com/SeSiTing/siti-cli/internal/shell"
	"github.com/spf13/cobra"
)

var aiSwitchCmd = &cobra.Command{
	Use:   "switch [provider]",
	Short: "切换 AI 服务商",
	Long: `默认切换当前 shell 的 Claude 与 Grok，不修改 Codex。
使用 --client codex 时全局修改 Codex 配置，同时影响 CLI、Desktop 与 IDE。`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		providers, err := availableProviders()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		completions := make([]cobra.Completion, 0, len(providers))
		for _, p := range providers {
			completions = append(completions, cobra.CompletionWithDesc(p.DisplayName(), clientSummary(p)))
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(c *cobra.Command, args []string) error {
		providers, err := availableProviders()
		if err != nil {
			return err
		}
		if len(providers) == 0 {
			return fmt.Errorf("未发现任何可用的 AI 服务商")
		}

		providerName := ""
		if len(args) > 0 {
			providerName = strings.ToUpper(args[0])
		} else {
			providerName, err = selectProvider(providers, config.CurrentProvider(providers))
			if err != nil || providerName == "" {
				return err
			}
		}
		p, ok := providers.Find(providerName)
		if !ok {
			return fmt.Errorf("服务商 '%s' 不存在，运行 'siti ai list' 查看可用服务商", strings.ToLower(providerName))
		}

		defaults := aiClients{claude: true, grok: p.SupportsGrok()}
		clients, err := parseAIClients(aiSwitchClient, defaults)
		if err != nil {
			return err
		}
		if err := preflightSwitch(p, clients); err != nil {
			return err
		}

		if clients.grok {
			if _, _, err := config.EnsureGrokProvider(p); err != nil {
				return fmt.Errorf("初始化 Grok 配置失败: %w", err)
			}
		}
		if clients.codex {
			path, backup, changed, err := applyCodexSwitch(p)
			if err != nil {
				return err
			}
			if changed {
				printErr("✓ Codex 已全局切换到 %s / %s", p.DisplayName(), p.CodexModel())
				printErr("  配置: %s", path)
				printErr("  备份: %s", backup)
			} else {
				printErr("✓ Codex 已是 %s / %s", p.DisplayName(), p.CodexModel())
			}
			printErr("! 影响 Codex CLI、Desktop 与 IDE；重启已运行的 Codex 后生效")
		}
		if clients.claude || clients.grok {
			applyShellSwitch(c, p, clients)
			printErr("✓ 当前 shell 已切换到 %s (%s)", p.DisplayName(), selectedShellClients(clients))
		}
		return nil
	},
}

func availableProviders() (config.ProviderList, error) {
	providers, err := config.ReadProviders()
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	skip := config.ReadSkipList()
	available := make(config.ProviderList, 0, len(providers))
	for _, p := range providers {
		if !p.IsSkipped(skip) {
			available = append(available, p)
		}
	}
	if len(available) == 0 && len(providers) > 0 {
		return nil, fmt.Errorf("所有服务商均在跳过列表中 (SITI_AI_SKIP)")
	}
	return available, nil
}

func preflightSwitch(p config.Provider, clients aiClients) error {
	if (clients.claude || clients.grok) && os.Getenv("SITI_WRAPPER") == "" && os.Getenv("SITI_EVAL_FILE") == "" {
		return fmt.Errorf("shell wrapper 未加载，当前 shell 无法切换\n请运行: eval \"$(siti init %s)\"", detectShell())
	}
	if clients.claude {
		if p.BaseURL() == "" {
			return fmt.Errorf("未找到 %s 的 Claude Base URL", p.DisplayName())
		}
		if err := requireEnvKey(p.AuthTokenVar); err != nil {
			return err
		}
	}
	if clients.grok {
		if !p.SupportsGrok() {
			return fmt.Errorf("服务商 '%s' 未配置 Grok Chat Completions 端点", p.DisplayName())
		}
		if err := requireEnvKey(p.GrokAuthTokenVar); err != nil {
			return err
		}
	}
	if clients.codex {
		if p.CodexUnsupportedReason != "" {
			return fmt.Errorf("服务商 '%s' 不能用于 Codex: %s", p.DisplayName(), p.CodexUnsupportedReason)
		}
		if !p.SupportsCodex() {
			return fmt.Errorf("服务商 '%s' 未配置 Responses-compatible Codex 端点", p.DisplayName())
		}
		has, err := config.HasCredential(p.DisplayName())
		if err != nil {
			return fmt.Errorf("读取系统凭证失败: %w", err)
		}
		if !has {
			if os.Getenv(p.CodexAuthTokenVar) != "" {
				return fmt.Errorf("Codex 全局切换需要系统凭证\n请先运行: siti ai credential import %s --from-env %s", p.DisplayName(), p.CodexAuthTokenVar)
			}
			return fmt.Errorf("未找到 %s\n请先设置该环境变量，再运行: siti ai credential import %s --from-env %s", p.CodexAuthTokenVar, p.DisplayName(), p.CodexAuthTokenVar)
		}
	}
	return nil
}

func requireEnvKey(name string) error {
	if name != "" && os.Getenv(name) != "" {
		return nil
	}
	if name == "" {
		name = "API_KEY"
	}
	return fmt.Errorf("未找到 %s\n请在 shell 配置中设置后重新加载；未修改任何客户端配置", name)
}

func applyShellSwitch(c *cobra.Command, p config.Provider, clients aiClients) {
	lines := make([]string, 0, 12)
	if clients.claude {
		lines = append(lines,
			exportResolved("ANTHROPIC_BASE_URL", p.BaseURLVar, p.BaseURLDefault),
			shell.ExportRef("ANTHROPIC_AUTH_TOKEN", p.AuthTokenVar),
		)
		if p.Model() == "" {
			lines = append(lines, shell.Unset(anthropicModelKeys...))
		} else {
			for _, key := range anthropicModelKeys {
				lines = append(lines, exportResolved(key, p.ModelVar, p.ModelDefault))
			}
		}
	}
	if clients.grok {
		lines = append(lines, shell.Export("SITI_GROK_MODEL_ID", config.GrokModelID(p)))
	}
	Eval(c, lines...)
}

func exportResolved(target, source, fallback string) string {
	if source != "" && os.Getenv(source) != "" {
		return shell.ExportRef(target, source)
	}
	return shell.Export(target, fallback)
}

func applyCodexSwitch(p config.Provider) (string, string, bool, error) {
	binary, err := exec.LookPath("siti")
	if err != nil {
		return "", "", false, fmt.Errorf("找不到 siti 可执行文件，无法配置 Codex 凭证助手")
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		return "", "", false, err
	}
	return config.ApplyCodexConfig(config.CodexProviderConfig{
		ProviderName: p.DisplayName(),
		DisplayName:  p.DisplayName() + " via siti-cli",
		BaseURL:      p.CodexBaseURL(),
		Model:        p.CodexModel(),
		AuthCommand:  binary,
		AuthArgs:     []string{"ai", "credential-helper", p.DisplayName()},
	})
}

func selectedShellClients(clients aiClients) string {
	names := make([]string, 0, 2)
	if clients.claude {
		names = append(names, "Claude")
	}
	if clients.grok {
		names = append(names, "Grok")
	}
	return strings.Join(names, " + ")
}

func clientSummary(p config.Provider) string {
	clients := []string{"claude"}
	if p.SupportsGrok() {
		clients = append(clients, "grok")
	}
	if p.SupportsCodex() {
		clients = append(clients, "codex")
	}
	return strings.Join(clients, ",")
}
