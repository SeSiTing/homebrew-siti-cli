package cmd

import (
	"fmt"

	"github.com/SeSiTing/siti-cli/internal/tunnel"
	"github.com/spf13/cobra"
)

var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "管理 SSH 本地端口转发",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var tunnelUpCmd = &cobra.Command{
	Use:   "up <profile>",
	Short: "后台启动 SSH tunnel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := tunnel.NewManager()
		if err != nil {
			return err
		}
		result, err := manager.Up(args[0])
		if err != nil {
			return err
		}
		if result.AlreadyRunning {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ tunnel %s 已在运行\n", args[0])
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ 已启动 tunnel %s\n", args[0])
		}
		printTunnelForwards(cmd, result.Status)
		return nil
	},
}

var tunnelDownCmd = &cobra.Command{
	Use:   "down <profile>",
	Short: "关闭后台 SSH tunnel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := tunnel.NewManager()
		if err != nil {
			return err
		}
		result, err := manager.Down(args[0])
		if err != nil {
			return err
		}
		if result.Stopped {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ 已关闭 tunnel %s\n", args[0])
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ tunnel %s 未运行\n", args[0])
		}
		return nil
	},
}

var tunnelStatusCmd = &cobra.Command{
	Use:   "status <profile>",
	Short: "查看 SSH tunnel 状态",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := tunnel.NewManager()
		if err != nil {
			return err
		}
		result, err := manager.Status(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s:\n", args[0])
		fmt.Fprintf(cmd.OutOrStdout(), "  target: %s\n", result.Profile.Target)
		if result.Running {
			fmt.Fprintln(cmd.OutOrStdout(), "  state: running")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "  state: stopped")
		}
		printTunnelForwards(cmd, result)
		return nil
	},
}

var tunnelListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可用的 tunnel profile",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := tunnel.NewManager()
		if err != nil {
			return err
		}
		profiles, err := manager.List()
		if err != nil {
			return err
		}
		if len(profiles) == 0 {
			dir, dirErr := tunnel.DefaultConfigDir()
			if dirErr != nil {
				return dirErr
			}
			fmt.Fprintf(cmd.OutOrStdout(), "未找到 tunnel profile: %s\n", dir)
			return nil
		}
		for _, profile := range profiles {
			fmt.Fprintln(cmd.OutOrStdout(), profile)
		}
		return nil
	},
}

func printTunnelForwards(cmd *cobra.Command, result tunnel.StatusResult) {
	for _, forward := range result.Forwards {
		state := "unreachable"
		if forward.Reachable {
			state = "reachable"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s: 127.0.0.1:%d -> %s:%d [%s]\n", forward.Name, forward.LocalPort, forward.RemoteHost, forward.RemotePort, state)
		if forward.URL != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", forward.URL)
		}
	}
}

func init() {
	tunnelCmd.AddCommand(tunnelUpCmd, tunnelDownCmd, tunnelStatusCmd, tunnelListCmd)
	rootCmd.AddCommand(tunnelCmd)
}
