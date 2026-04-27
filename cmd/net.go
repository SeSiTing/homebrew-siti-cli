package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var netCmd = &cobra.Command{
	Use:   "net",
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
	rootCmd.AddCommand(netCmd)
}
