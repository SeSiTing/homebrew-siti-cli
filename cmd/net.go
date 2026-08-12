package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/SeSiTing/siti-cli/internal/network"
	"github.com/spf13/cobra"
)

var netCmd = &cobra.Command{
	Use:   "net",
	Short: "管理 macOS 网络配置",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var netApplyCmd = &cobra.Command{
	Use:   "apply <profile>",
	Short: "应用指定的固定网络配置",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := network.NewManager()
		if err != nil {
			return err
		}
		result, err := manager.Apply(args[0])
		if err != nil {
			return err
		}
		if result.AlreadyApplied {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ network profile %s 已生效\n", result.State.Profile)
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ 已应用 network profile %s\n", result.State.Profile)
		fmt.Fprintf(cmd.OutOrStdout(), "  service: %s (%s)\n", result.State.Service, result.State.Device)
		if result.State.SSID != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Wi-Fi: %s\n", result.State.SSID)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  IPv4: %s/%s\n", result.State.IPv4.Address, result.State.IPv4.SubnetMask)
		fmt.Fprintf(cmd.OutOrStdout(), "  gateway: %s\n", result.State.IPv4.Gateway)
		fmt.Fprintf(cmd.OutOrStdout(), "  DNS: %s\n", strings.Join(result.State.DNS, ", "))
		return nil
	},
}

var netResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "恢复 DHCP 和自动 DNS",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := network.NewManager()
		if err != nil {
			return err
		}
		result, err := manager.Reset()
		if err != nil {
			return err
		}
		if !result.Changed {
			fmt.Fprintln(cmd.OutOrStdout(), "✓ 当前没有 siti 管理的网络配置")
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), "✓ 已恢复自动网络配置")
		fmt.Fprintf(cmd.OutOrStdout(), "  service: %s (%s)\n", result.Service, result.State.Device)
		fmt.Fprintln(cmd.OutOrStdout(), "  IPv4: DHCP")
		if result.Live.Address != "" {
			address := result.Live.Address
			if result.Live.SubnetMask != "" {
				address += "/" + result.Live.SubnetMask
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  address: %s\n", address)
		}
		if result.Live.Gateway != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  gateway: %s\n", result.Live.Gateway)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "  DNS: Automatic")
		return nil
	},
}

var netStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前 siti 管理的网络配置",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := network.NewManager()
		if err != nil {
			return err
		}
		result, err := manager.Status()
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "当前网络配置:")
		if !result.Active {
			fmt.Fprintln(cmd.OutOrStdout(), "  managed: none")
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  profile: %s\n", result.State.Profile)
		fmt.Fprintf(cmd.OutOrStdout(), "  service: %s (%s)\n", result.Service, result.State.Device)
		if result.State.SSID != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Wi-Fi: %s\n", result.State.SSID)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  mode: %s\n", result.Live.Mode)
		fmt.Fprintf(cmd.OutOrStdout(), "  IPv4: %s/%s\n", result.Live.Address, result.Live.SubnetMask)
		fmt.Fprintf(cmd.OutOrStdout(), "  gateway: %s\n", result.Live.Gateway)
		fmt.Fprintf(cmd.OutOrStdout(), "  DNS: %s\n", strings.Join(result.Live.DNS, ", "))
		return nil
	},
}

var netListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可用的 network profile",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := network.DefaultConfigDir()
		if err != nil {
			return err
		}
		profiles, err := network.ListProfiles(dir)
		if err != nil {
			return err
		}
		active, hasActive, err := network.ReadActive(dir)
		if err != nil {
			return err
		}
		if len(profiles) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "未找到 network profile: %s\n", dir)
			return nil
		}
		for _, profile := range profiles {
			if hasActive && active.Profile == profile {
				fmt.Fprintf(cmd.OutOrStdout(), "%s [active]\n", profile)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), profile)
			}
		}
		return nil
	},
}

var netCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "检查网络连接状态（ping baidu/google/github）",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		targets := []string{"baidu.com", "google.com", "github.com"}
		for _, target := range targets {
			fmt.Printf("→ ping %s\n", target)
			c := exec.Command("ping", "-c", "2", target)
			c.Stdout = cmd.OutOrStdout()
			c.Stderr = cmd.ErrOrStderr()
			_ = c.Run()
			fmt.Println()
		}
	},
}

func init() {
	netCmd.AddCommand(netApplyCmd, netResetCmd, netStatusCmd, netListCmd, netCheckCmd)
	rootCmd.AddCommand(netCmd)
}
