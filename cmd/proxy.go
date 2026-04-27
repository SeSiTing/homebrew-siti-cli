package cmd

import (
	"fmt"

	"github.com/SeSiTing/siti-cli/internal/shell"
	"github.com/spf13/cobra"
)

const (
	proxyHost = "127.0.0.1"
	proxyPort = "7890"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "管理终端代理设置",
}

var proxyOnCmd = &cobra.Command{
	Use:   "on",
	Short: "开启终端代理",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		httpProxy := fmt.Sprintf("http://%s:%s", proxyHost, proxyPort)
		socksProxy := fmt.Sprintf("socks5://%s:%s", proxyHost, proxyPort)
		printErr("✓ 终端代理已开启 (%s:%s)", proxyHost, proxyPort)
		Eval(c,
			shell.Export("http_proxy", httpProxy),
			shell.Export("HTTP_PROXY", httpProxy),
			shell.Export("https_proxy", httpProxy),
			shell.Export("HTTPS_PROXY", httpProxy),
			shell.Export("all_proxy", socksProxy),
			shell.Export("ALL_PROXY", socksProxy),
		)
		return nil
	},
}

var proxyOffCmd = &cobra.Command{
	Use:   "off",
	Short: "关闭终端代理",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		printErr("✗ 终端代理已关闭")
		Eval(c,
			shell.Unset("http_proxy", "HTTP_PROXY"),
			shell.Unset("https_proxy", "HTTPS_PROXY"),
			shell.Unset("all_proxy", "ALL_PROXY"),
		)
		return nil
	},
}

var proxyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前代理状态",
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		httpVal := firstNonEmpty(
			lookupEnv("http_proxy"),
			lookupEnv("HTTP_PROXY"),
		)
		httpsVal := firstNonEmpty(
			lookupEnv("https_proxy"),
			lookupEnv("HTTPS_PROXY"),
		)
		allVal := firstNonEmpty(
			lookupEnv("all_proxy"),
			lookupEnv("ALL_PROXY"),
		)

		fmt.Println("当前代理状态:")
		if httpVal != "" {
			fmt.Println("  ✓ 代理已开启")
			fmt.Println("  http_proxy: ", httpVal)
			fmt.Println("  https_proxy:", httpsVal)
			fmt.Println("  all_proxy:  ", allVal)
		} else {
			fmt.Println("  ✗ 代理未开启")
		}

		if v := firstNonEmpty(lookupEnv("no_proxy"), lookupEnv("NO_PROXY")); v != "" {
			fmt.Println("  no_proxy:   ", v)
		}
	},
}

func init() {
	proxyCmd.AddCommand(proxyOnCmd, proxyOffCmd, proxyStatusCmd)
	rootCmd.AddCommand(proxyCmd)
}
