package cmd

import (
	"fmt"
	"strings"

	"github.com/SeSiTing/siti-cli/internal/config"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var anthropicModelKeys = []string{
	"ANTHROPIC_MODEL",
	"ANTHROPIC_DEFAULT_SONNET_MODEL",
	"ANTHROPIC_DEFAULT_OPUS_MODEL",
	"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	"ANTHROPIC_REASONING_MODEL",
}

var grokEnvKeys = []string{
	"SITI_GROK_MODEL_ID",
}

type aiClients struct {
	claude bool
	grok   bool
	codex  bool
}

var (
	aiSwitchClient  string
	aiClearClient   string
	aiCurrentClient string
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "管理 AI API 配置切换",
}

var aiSetupCmd = &cobra.Command{
	Use:       "setup [grok]",
	Short:     "初始化 Grok Build 的 siti 模型入口",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"grok"},
	RunE: func(c *cobra.Command, args []string) error {
		if strings.ToLower(args[0]) != "grok" {
			return fmt.Errorf("暂不支持初始化: %s\n支持: grok", args[0])
		}
		providers, err := availableProviders()
		if err != nil {
			return err
		}
		count := 0
		for _, p := range providers {
			if !p.SupportsGrok() {
				continue
			}
			path, _, err := config.EnsureGrokProvider(p)
			if err != nil {
				return fmt.Errorf("初始化 %s Grok 配置失败: %w", p.DisplayName(), err)
			}
			printErr("✓ Grok 模型入口已就绪: %s (%s)", config.GrokModelID(p), path)
			count++
		}
		if count == 0 {
			return fmt.Errorf("没有支持 Grok 的可用服务商")
		}
		return nil
	},
}

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

func parseAIClients(value string, defaults aiClients) (aiClients, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default":
		return defaults, nil
	case "claude":
		return aiClients{claude: true}, nil
	case "grok":
		return aiClients{grok: true}, nil
	case "codex":
		return aiClients{codex: true}, nil
	case "all":
		return aiClients{claude: true, grok: true, codex: true}, nil
	default:
		return aiClients{}, fmt.Errorf("不支持的客户端 '%s'，可选: claude, grok, codex, all", value)
	}
}

func init() {
	aiSwitchCmd.Flags().StringVar(&aiSwitchClient, "client", "default", "切换客户端: claude, grok, codex, all")
	aiClearCmd.Flags().StringVar(&aiClearClient, "client", "default", "清除客户端: claude, grok, codex, all")
	aiCurrentCmd.Flags().StringVar(&aiCurrentClient, "client", "all", "查看客户端: claude, grok, codex, all")
	aiCmd.AddCommand(
		aiSwitchCmd,
		aiSetupCmd,
		aiListCmd,
		aiCurrentCmd,
		aiDoctorCmd,
		aiTestCmd,
		aiClearCmd,
		aiCredentialCmd,
		aiCredentialHelperCmd,
	)
	rootCmd.AddCommand(aiCmd)
}
