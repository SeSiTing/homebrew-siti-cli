package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SeSiTing/siti-cli/internal/tunnel"
	"github.com/charmbracelet/x/ansi"
)

func TestWriteTunnelForwardsReadablePlainText(t *testing.T) {
	result := tunnel.StatusResult{Forwards: []tunnel.ForwardStatus{
		{
			Forward:   tunnel.Forward{Name: "openclaw", LocalPort: 19010, RemoteHost: "127.0.0.1", RemotePort: 9010, URL: "http://127.0.0.1:19010/"},
			Reachable: true,
		},
		{
			Forward: tunnel.Forward{Name: "hermes", LocalPort: 19119, RemoteHost: "127.0.0.1", RemotePort: 9119, URL: "http://127.0.0.1:19119/"},
		},
	}}

	var output bytes.Buffer
	writeTunnelForwards(&output, result, false)

	want := "  openclaw\n" +
		"    打开: http://127.0.0.1:19010/\n" +
		"    转发: 127.0.0.1:19010 → 127.0.0.1:9010\n" +
		"    状态: SSH 转发就绪\n\n" +
		"  hermes\n" +
		"    打开: http://127.0.0.1:19119/\n" +
		"    转发: 127.0.0.1:19119 → 127.0.0.1:9119\n" +
		"    状态: 本地端口未监听\n\n"
	if output.String() != want {
		t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", want, output.String())
	}
}

func TestFormatTerminalURLHyperlink(t *testing.T) {
	url := "http://127.0.0.1:19010/"
	got := formatTerminalURL(url, true)
	want := ansi.SetHyperlink(url) + url + ansi.ResetHyperlink()
	if got != want {
		t.Fatalf("formatTerminalURL() = %q, want %q", got, want)
	}
	if !strings.Contains(got, url) {
		t.Fatalf("hyperlink must retain visible URL: %q", got)
	}
}
