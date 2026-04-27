package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "显示内网和公网 IP 地址",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		lan := "(未获取到)"
		if out, err := exec.Command("ipconfig", "getifaddr", "en0").Output(); err == nil {
			lan = strings.TrimSpace(string(out))
		}
		fmt.Printf("LAN: %s\n", lan)
		fmt.Printf("WAN: %s\n", lookupPublicIP())
	},
}

// lookupPublicIP tries multiple endpoints; returns the first plain-text IP it gets.
func lookupPublicIP() string {
	endpoints := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://ipinfo.io/ip",
	}
	client := &http.Client{Timeout: 5 * time.Second}
	for _, url := range endpoints {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64))
		resp.Body.Close()
		ip := strings.TrimSpace(string(body))
		if ip != "" && !strings.ContainsAny(ip, "<> ") {
			return ip
		}
	}
	return "(获取失败)"
}

func init() {
	rootCmd.AddCommand(ipCmd)
}
