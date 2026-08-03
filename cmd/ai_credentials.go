package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/SeSiTing/siti-cli/internal/config"
	"github.com/spf13/cobra"
)

var aiCredentialImportEnv string

var aiCredentialCmd = &cobra.Command{
	Use:   "credential",
	Short: "管理 AI 系统凭证",
	Args:  cobra.NoArgs,
}

var aiCredentialImportCmd = &cobra.Command{
	Use:   "import <provider>",
	Short: "从环境变量导入 API Key 到系统凭证库",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		providers, err := config.ReadProviders()
		if err != nil {
			return fmt.Errorf("读取配置失败: %w", err)
		}
		p, ok := providers.Find(args[0])
		if !ok {
			return fmt.Errorf("服务商 '%s' 不存在", args[0])
		}
		envName := aiCredentialImportEnv
		if envName == "" {
			envName = p.CodexAuthTokenVar
			if envName == "" {
				envName = p.AuthTokenVar
			}
		}
		token := os.Getenv(envName)
		if token == "" {
			return fmt.Errorf("未找到环境变量 %s；未写入系统凭证库", envName)
		}
		if err := config.SetCredential(p.DisplayName(), token); err != nil {
			return fmt.Errorf("写入系统凭证失败: %w", err)
		}
		printErr("✓ 已将 %s 从 %s 导入系统凭证库", p.DisplayName(), envName)
		return nil
	},
}

var aiCredentialStatusCmd = &cobra.Command{
	Use:   "status [provider]",
	Short: "查看系统凭证状态（不显示密钥）",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		providers, err := config.ReadProviders()
		if err != nil {
			return fmt.Errorf("读取配置失败: %w", err)
		}
		if len(args) == 1 {
			p, ok := providers.Find(args[0])
			if !ok {
				return fmt.Errorf("服务商 '%s' 不存在", args[0])
			}
			providers = config.ProviderList{p}
		}
		for _, p := range providers {
			has, err := config.HasCredential(p.DisplayName())
			if err != nil {
				return fmt.Errorf("读取 %s 系统凭证失败: %w", p.DisplayName(), err)
			}
			state := "未保存"
			if has {
				state = "已保存"
			}
			fmt.Printf("%-15s %s\n", p.DisplayName(), state)
		}
		return nil
	},
}

var aiCredentialRemoveCmd = &cobra.Command{
	Use:   "remove <provider>",
	Short: "从系统凭证库删除 API Key",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		provider := strings.ToLower(args[0])
		if err := config.DeleteCredential(provider); err != nil {
			return fmt.Errorf("删除系统凭证失败: %w", err)
		}
		printErr("✓ 已删除 %s 系统凭证", provider)
		return nil
	},
}

var aiCredentialHelperCmd = &cobra.Command{
	Use:    "credential-helper <provider>",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		token, err := config.GetCredential(args[0])
		if err != nil {
			return fmt.Errorf("读取系统凭证失败: %w", err)
		}
		fmt.Print(token)
		return nil
	},
}

func init() {
	aiCredentialImportCmd.Flags().StringVar(&aiCredentialImportEnv, "from-env", "", "读取 API Key 的环境变量名")
	aiCredentialCmd.AddCommand(aiCredentialImportCmd, aiCredentialStatusCmd, aiCredentialRemoveCmd)
}
