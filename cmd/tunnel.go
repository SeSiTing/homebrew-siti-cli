package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SeSiTing/siti-cli/internal/tunnel"
	"github.com/charmbracelet/x/ansi"
	charmterm "github.com/charmbracelet/x/term"
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
		fmt.Fprintln(cmd.OutOrStdout())
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
		fmt.Fprintln(cmd.OutOrStdout(), args[0])
		fmt.Fprintf(cmd.OutOrStdout(), "  SSH 目标: %s\n", result.Profile.Target)
		if result.Running {
			fmt.Fprintln(cmd.OutOrStdout(), "  状态: 运行中")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "  状态: 已停止")
		}
		fmt.Fprintln(cmd.OutOrStdout())
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
	writer := cmd.OutOrStdout()
	writeTunnelForwards(writer, result, terminalHyperlinksEnabled(writer))
}

func writeTunnelForwards(writer io.Writer, result tunnel.StatusResult, hyperlinks bool) {
	for _, forward := range result.Forwards {
		state := "本地端口未监听"
		if forward.Reachable {
			state = "SSH 转发就绪"
		}
		fmt.Fprintf(writer, "  %s\n", forward.Name)
		if forward.URL != "" {
			fmt.Fprintf(writer, "    打开: %s\n", formatTerminalURL(forward.URL, hyperlinks))
		}
		fmt.Fprintf(writer, "    转发: 127.0.0.1:%d → %s:%d\n", forward.LocalPort, forward.RemoteHost, forward.RemotePort)
		fmt.Fprintf(writer, "    状态: %s\n", state)
		fmt.Fprintln(writer)
	}
}

func formatTerminalURL(url string, hyperlinks bool) string {
	if !hyperlinks {
		return url
	}
	return ansi.SetHyperlink(url) + url + ansi.ResetHyperlink()
}

func terminalHyperlinksEnabled(writer io.Writer) bool {
	if os.Getenv("CI") != "" || strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	file, ok := writer.(interface{ Fd() uintptr })
	return ok && charmterm.IsTerminal(file.Fd())
}

func init() {
	tunnelCmd.AddCommand(tunnelUpCmd, tunnelDownCmd, tunnelStatusCmd, tunnelListCmd)
	rootCmd.AddCommand(tunnelCmd)
}
